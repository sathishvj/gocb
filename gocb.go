package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	help := flag.Bool("h", false, "to display this usage listing")
	run := flag.Bool("r", false, "run the file once if build was ok")
	test := flag.Bool("t", false, "execute tests once if build was ok (untried)")
	silent := flag.Bool("s", false, "fairly silent in output")
	interval := flag.Int("i", 1, "Polling interval (in seconds)")
	//params := flag.String("p", "", "Parameters to be sent to program when run")
	flag.Parse()
	//anyFlags := help

	if *help {
		usage()
		return
	}

	if len(flag.Args()) == 0 {
		fmt.Println("Error! No arguments given.\n")
		usage()
		return
	}

	var watch string
	if len(flag.Args()) == 0 {
		watch = "."
	} else {
		watch = flag.Args()[0]
	}
	fmt.Println("Watching:", watch)

	for {
		//isChanged := fi.ModTime().Sub(prevMod) > 0
		changed, changes, err := isChanged(watch)
		if err != nil {
			fmt.Println("Error: ", err.Error(), "\nExiting.")
			return
		}

		if changed {
			if len(changes) > 3 {
				changes = append(changes[0:2], "... "+strconv.Itoa(len(changes)))
			}
			if !*silent {
				fmt.Println("  --->> File change detected: ", strings.Join(changes, ","))
				fmt.Println("     >> starting build")
			} else {
				fmt.Println("  --->>")
			}
			//prevMod = fi.ModTime()
			buildErr := exe(watch, "build")
			if buildErr && !*silent {
				fmt.Println("     >> build finished with errors")
			} else if !*silent {
				fmt.Println("     >> build ok")
			}

			if !buildErr && *test {
				if !*silent {
					fmt.Println("     >> test starting")
				}
				exe(watch, "test")
				if !*silent {
					fmt.Println("     >> test finished")
				}
			}

			if !buildErr && *run {
				if !*silent {
					fmt.Println("     >> run starting")
				}
				exe(watch, "run")
				if !*silent {
					fmt.Println("     >> run finished")
				}
			}
		}
		time.Sleep(time.Duration(*interval) * time.Second)
	}
}

var prevFileMod time.Time
var prevDirMod map[string]time.Time
var dirChanges []string

//returns true if there are changes
//returns an array of files that have changes
//return a non nil error if there was an error
func isChanged(watch string) (bool, []string, error) {
	fi, err := os.Stat(watch)
	if err != nil {
		return false, nil, err
	}
	isDir := fi.IsDir()
	if isDir {
		if prevDirMod == nil {
			prevDirMod = make(map[string]time.Time)
		}

		visit := func(path string, info os.FileInfo, e error) error {
			if !info.IsDir() {
				if strings.HasSuffix(strings.ToLower(path), ".go") {
					t, ok := prevDirMod[path]
					if !ok {
						prevDirMod[path] = info.ModTime()
						dirChanges = append(dirChanges, path)
					} else {
						if info.ModTime().Sub(t) > 0 {
							prevDirMod[path] = info.ModTime()
							dirChanges = append(dirChanges, path)
						}
					}
				}
			}
			return nil
		}

		dirChanges = dirChanges[0:0]
		err := filepath.Walk(watch, visit)
		if err != nil {
			return false, nil, err
		}

		if len(dirChanges) > 0 {
			return true, dirChanges, nil
		}

	} else {
		if fi.ModTime().Sub(prevFileMod) > 0 {
			prevFileMod = fi.ModTime()
			return true, []string{watch}, nil
		}
	}
	return false, nil, nil
}

//returns true if there was an output on std error
func exe(watch string, tool string) bool {

	cmd := exec.Command("go", tool, watch)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error getting StdoutPipe")
		panic(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("Error getting StdErrPipe")
		panic(err)
	}

	outPipe := bufio.NewReader(stdout)
	errPipe := bufio.NewReader(stderr)
	outCh := make(chan string)
	errCh := make(chan string)

	err = cmd.Start()
	if err != nil {
		fmt.Println("Error starting cmd")
		panic(err)
	}
	go getStdOutput(outCh, outPipe)
	go getStdOutput(errCh, errPipe)
	hasErrOutput := false

outside:
	for {
		select {
		case s, ok := <-outCh:
			if !ok {
				break outside
			}
			//fmt.Println("From Out:", s)
			fmt.Println(s)
		case s, ok := <-errCh:
			if !ok {
				break outside
			}
			//fmt.Println("Err!", s)
			fmt.Println(s)

			// This should be a compile error.  (I think.)
			hasErrOutput = true
		}
	}
	//fmt.Println("Finished")

	err = cmd.Wait()
	if err != nil {
		// Typically a compile error. Don't print any message.
		//fmt.Println("Waiting error", err)
		return hasErrOutput
	}
	return hasErrOutput
}

func getStdOutput(c chan string, p *bufio.Reader) {
	for {
		line, _, err := p.ReadLine()
		if err != nil {
			if err != io.EOF {
				fmt.Println("ReadLine error:", err)
				//return
			}
			break
		}
		//fmt.Println("Got line:", string(line))
		c <- string(line)
	}
	//fmt.Println("Closing c")
	close(c)
}

func usage() {
	pgm := os.Args[0]
	fmt.Printf("%s will watch for the file change and build it.  Usage:\n", pgm)
	flag.PrintDefaults()
	fmt.Printf("\n%s hello.go            \n", pgm)
	fmt.Printf("%s -r hello.go         #Also run the file on successful build.\n", pgm)
	return
}

func write(m map[string]map[string]string, f string) error {
	var s string
	for k, v := range m {
		for k1, v1 := range v {
			s = s + k + "=" + k1 + "=" + v1 + "\n"
		}
	}
	err := ioutil.WriteFile(f, []byte(s), 0644)
	if err != nil {
		return err
	}
	return nil
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
)

func main() {
	help := flag.Bool("h", false, "to display this usage listing")
	run := flag.Bool("r", false, "run the file once if build was ok")
	test := flag.Bool("t", false, "run the file once if build was ok")
	interval := flag.Int("i", 1, "Polling interval (in seconds)")
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

	watchFile := flag.Args()[0]
	fmt.Println("Watching file:", watchFile)

	var prevMod time.Time
	for {
		fi, err := os.Stat(watchFile)
		if err != nil {
			fmt.Println("Error checking file: ", err)
			return
		}
		modTime := fi.ModTime()
		if modTime.Sub(prevMod) > 0 {
			fmt.Println("   >> File change detected: ", watchFile)
			fmt.Println("   >> starting build")
			prevMod = modTime
			buildErr := exe(watchFile, "build")
			if buildErr {
				fmt.Println("   >> build finished with errors")
			} else {
				fmt.Println("   >> build ok")
			}

			if !buildErr && *test {
				fmt.Println("   >> test starting")
				exe(watchFile, "test")
				fmt.Println("   >> test finished")
			}

			if !buildErr && *run {
				fmt.Println("   >> run starting")
				exe(watchFile, "run")
				fmt.Println("   >> run finished")
			}
		}
		time.Sleep(time.Duration(*interval) * time.Second)
	}

}

//returns true if there was an output on std error
func exe(watchFile string, tool string) bool {

	cmd := exec.Command("go", tool, watchFile)
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

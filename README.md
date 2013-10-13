gocb
===

## Build (and run) a go file (or directory) when it changes
Repeatedly doing "go build" or "go run" on a file is tedious.  gocb continuosly watches for changes to a go file (or a directory), and if it has changed, it automatically runs a "go build" on it.  If you have given the -r option, it will also do a "go run" if the build was successful.

The best way to use it is to have a second terminal window/pane where gocb is running.  In your editing window, as soon as you save changes, you will see the build result in the gocb window.

## Installation
```
go get github.com/sathishvj/gocb
```

> This will pull down the source and install the gocb command in $GOPATH/bin

## Running gocb
Make sure that your PATH contains $GOPATH/bin.  Then use gocb from any directory.

```
gocb -r hello.go 
```
> this will watch hello.go for changes and run a build if there are any.  Since the -r option is present, it will also run the file if the build is successful.


```
gocb -s -i 5 hello.go 
```
> Poll at an interval of 5 seconds.  Default is 1 second.
> Also, -s option causes the output to be fairly silent in messages from gocb.

```
gocb .
```
> Watch the current directory for any changes, and do a build if there are any.

## Help
```
gocb -h
```

## Testing
Checked this only on Mac OS X with single files and with directories with a couple of files. If you find any bugs, please raise an issue for this project.  Thank you.

## ToDo
* Add option to set a specific go binary.
* In execution mode, add option to quit when q is pressed, toggle run when r is pressed, toggle test when t is pressed, increase/decrease time interval when +/- is pressed.

// Command fsnotify provides example usage of the fsnotify library.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var usage = `
war is a tiny cross platform file watcher

-d <directory>
-r <command>

Example:

    war -d . -r "echo building" -r make -r myapp.exe

    war -d /tmp -r "echo hello" -r "echo world"

    The above will watch the current direcotry (.) and
    run the three commands given with -r in order if
    the prior one was success.

    
`

func exit(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, filepath.Base(os.Args[0])+": "+format+"\n", a...)
	fmt.Print("\n" + usage)
	os.Exit(1)
}

func help() {
	fmt.Printf("%s [command] [arguments]\n\n", filepath.Base(os.Args[0]))
	fmt.Print(usage)
	os.Exit(0)
}

// Print line prefixed with the time (a bit shorter than log.Print; we don't
// really need the date and ms is useful here).
func printTime(s string, args ...interface{}) {
	fmt.Printf(time.Now().Format("15:04:05.0000")+" "+s+"\n", args...)
}

func main() {
	if len(os.Args) == 1 {
		help()
	}
	// Always show help if -h[elp] appears anywhere before we do anything else.
	for _, f := range os.Args[1:] {
		switch f {
		case "help", "-h", "-help", "--help":
			help()
		}
	}

	args := os.Args[1:]
	dedup(args...)
}

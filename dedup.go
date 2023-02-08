package main


import (
	"math"
	"sync"
	"time"
	"os/exec"
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// Depending on the system, a single "write" can generate many Write events; for
// example compiling a large Go program can generate hundreds of Write events on
// the binary.
//
// The general strategy to deal with this is to wait a short time for more write
// events, resetting the wait period for every new event.
var runCommands = []string{}
func dedup(args ...string) {
	if len(args) < 4 {
		exit("must specify at least one -d path -a action")
	}

	// Create a new watcher.
	w, err := fsnotify.NewWatcher()
	if err != nil {
		exit("creating a new watcher: %s", err)
	}
	defer w.Close()


	// Start listening for events.
	go dedupLoop(w)

	// Add all args from the commandline.
	for k, p := range args {
		if p == "-d" {
			err = w.Add(args[k+1])
			if err != nil {
				exit("%q: %s", p, err)
			}
		}
		if p == "-r" {
			runCommands = append(runCommands, args[k+1])
		}
			
	}
	

	printTime("watching; press ^C to exit")
	<-make(chan struct{}) // Block forever
}

func dedupLoop(w *fsnotify.Watcher) {
	var (
		// Wait 250ms for new events; each new event resets the timer.
		waitFor = 250 * time.Millisecond

		// Keep track of the timers, as path → timer.
		mu     sync.Mutex
		timers = make(map[string]*time.Timer)

		// Callback we run.
		printEvent = func(e fsnotify.Event) {
			printTime(e.String())

			good := true
			for _, r := range runCommands {
				if good {
					cmd := exec.Command(strings.Split(r, " ")[0], strings.Split(r, " ")[1:]...)
					out, err := cmd.CombinedOutput()
					fmt.Println(string(out))
					if err != nil {
						good = false
						fmt.Println("failed to run", err)
					}
					
				}
			}

			// Don't need to remove the timer if you don't have a lot of files.
			mu.Lock()
			delete(timers, e.Name)
			mu.Unlock()
		}
	)

	for {
		select {
		// Read from Errors.
		case err, ok := <-w.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			printTime("ERROR: %s", err)
		// Read from Events.
		case e, ok := <-w.Events:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}

			// We just want to watch for file creation, so ignore everything
			// outside of Create and Write.
			if !e.Has(fsnotify.Create) && !e.Has(fsnotify.Write) {
				continue
			}

			// Get timer.
			mu.Lock()
			t, ok := timers[e.Name]
			mu.Unlock()

			// No timer yet, so create one.
			if !ok {
				t = time.AfterFunc(math.MaxInt64, func() { printEvent(e) })
				t.Stop()

				mu.Lock()
				timers[e.Name] = t
				mu.Unlock()
			}

			// Reset the timer for this path, so it will start from 250ms again.
			t.Reset(waitFor)
		}
	}
}

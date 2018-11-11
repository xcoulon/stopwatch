package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/chzyer/readline"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	io.WriteString(w, completer.Tree("    "))
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("start"),
	readline.PcItem("stop"),
	readline.PcItem("exit"),
	readline.PcItem("status"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

// initializes the level for the logger, using the optional '-debug' flag to activate the logs in 'debug' level.
// Other tests must import this 'test' package even if unused, using:
// import _ "github.com/bytesparadise/libasciidoc/pkg/log"
func init() {
	customFormatter := new(log.TextFormatter)
	customFormatter.DisableLevelTruncation = true
	customFormatter.DisableTimestamp = true
	customFormatter.ForceColors = true
	log.SetFormatter(customFormatter)
}

func main() {
	l, err := readline.NewEx(&readline.Config{
		Prompt:       "\033[31m»\033[0m ",
		HistoryFile:  "/tmp/vatriathlon-stopwatch.tmp",
		AutoComplete: completer,
		// InterruptPrompt: "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}

	defer func() {
		fmt.Println("exiting")
		os.Exit(0)
	}()
	defer func() {
		fmt.Println("closing readline")
		l.Close()
	}()
	defer func() {
		fmt.Println("TODO: closing db connection")
		l.Close()
	}()

	log.SetOutput(l.Stderr())
	// main loop that catches the commands
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			log.Warn("ignoring input...")
			continue
		} else if err != nil {
			log.Warn("ignoring error of type '%T'\n", err)
		}
		line = strings.TrimSpace(line)
		switch {
		case line == "start":
			log.Warn("started ⏱")
			now := time.Now()
			startTime = &now
			go func() {
				for range time.Tick(time.Second) {
					duration, _ := currentDuration()
					log.Infof("%s", duration)
				}
			}()
		case line == "stop":
			println("stopped ⏱")
		case line == "help":
			usage(l.Stderr())
		case line == "exit":
			log.Error("need a prompt to confirm! (need to 'stop' first?)") //TODO
			runtime.Goexit()
		default:
			log.Println("you said:", strconv.Quote(line))
		}
	}
}

var startTime *time.Time

func currentDuration() (time.Time, error) {
	if startTime == nil {
		return time.Time{}, errors.New("stopwach did not start yet")
	}
	now := time.Now()
	duration := now.Sub(*startTime)
	return time.Parse("03:04:05", strconv.Itoa(int(duration)))
}

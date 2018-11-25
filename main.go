package main

import (
	"fmt"
	"os"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/vatriathlon/stopwatch/service"

	"github.com/vatriathlon/stopwatch/configuration"

	"github.com/vatriathlon/stopwatch/cmd"

	"github.com/chzyer/readline"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

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
	config, err := configuration.New()
	if err != nil {
		panic(err)
	}
	l, err := readline.NewEx(&readline.Config{
		Prompt:      "\033[31m»\033[0m ",
		HistoryFile: "/tmp/vatriathlon-stopwatch.tmp",
		// AutoComplete: cmd.Completer,
		// InterruptPrompt: "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
		FuncFilterInputRune: func(r rune) (rune, bool) {
			switch r {
			// block CtrlZ feature
			case readline.CharCtrlZ:
				return r, false
			}
			return r, true
		},
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

	var db *gorm.DB
	for {
		db, err = gorm.Open("postgres", config.GetPostgresConfigString())
		if err != nil {
			db.Close()
			log.Errorf("ERROR: Unable to open connection to database %v", err)
			log.Infof("Retrying to connect in %v...", config.GetPostgresConnectionRetrySleep())
			time.Sleep(config.GetPostgresConnectionRetrySleep())
		} else {
			defer db.Close()
			break
		}
	}
	db.LogMode(config.IsDBLogsEnabled())

	svc := service.NewApplicationService(db)
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
		cmd.Execute(line, svc, l)
	}
}

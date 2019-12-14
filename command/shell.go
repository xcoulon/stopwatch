package command

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vatriathlon/stopwatch/pkg/configuration"
	"github.com/vatriathlon/stopwatch/pkg/connection"
	"github.com/vatriathlon/stopwatch/pkg/service"

	"github.com/mitchellh/cli"
	"github.com/peterh/liner"
)

// ShellCommand the command to run the shell
type ShellCommand struct {
	ui     cli.Ui
	flags  *flag.FlagSet
	help   string
	raceID int
}

var (
	history = filepath.Join("./tmp/liner_example_history")
	cmds    []string
)

var _ cli.Command = (*ShellCommand)(nil)

// NewShellCommand returns a new command to run the shell
func NewShellCommand(ui cli.Ui) *ShellCommand {
	c := &ShellCommand{
		ui: ui,
	}
	c.init()
	return c
}

func (c *ShellCommand) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.IntVar(&c.raceID, "race", 0, "The ID of the race")
}

// Run executes the command
func (c *ShellCommand) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	if c.raceID == 0 {
		c.ui.Error("the --race parameter must be specified")
		return 1
	}

	l := liner.NewLiner()
	defer l.Close()

	l.SetCtrlCAborts(true)
	if f, err := os.Open(history); err == nil {
		l.ReadHistory(f)
		f.Close()
	}

	config, err := configuration.New()
	if err != nil {
		c.ui.Error(fmt.Sprintf("error while loading the configuration: %s", err.Error()))
		return 1
	}
	db, err := connection.New(config)
	if err != nil {
		c.ui.Error(fmt.Sprintf("error while loading the configuration: %s", err.Error()))
		return 1
	}
	defer func() {
		db.Close()
	}()
	svc := service.NewRaceService(db)

	race, err := svc.GetRace(c.raceID)
	if err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	c.ui.Info("********************************************************")
	c.ui.Info("type '\\?' for help")
	c.ui.Info("********************************************************")

loop:
	for {
		cmd, err := l.Prompt(fmt.Sprintf("%s > ", race.Name))
		if err != nil {
			c.ui.Error(fmt.Sprintf("error occurred: %s", err.Error()))
			continue
		}
		cmd = strings.TrimSpace(cmd)
		l.AppendHistory(cmd)

		if cmd == "\\?" {
			c.ui.Info(`Available commands are:
	\?	displays help
	\r 	displays the current race
	\t	lists the teams in the current race
	\q	quits
	\s	starts the race with the given ID
	\d	deletes the last lap of the team with the given ID
	n	adds a lap to the team with the given ID
			
			`)
		} else if cmd == "\\q" {
			// exit the shell
			break loop
			// } else if cmd == "\\r" { // not implemented yet
			// } else if cmd == "\\t" { // not implemented yet
		} else if cmd == "\\s" {
			race, err = svc.StartRace(race.ID)
			if err != nil {
				c.ui.Error(err.Error())
				continue
			}
		} else if cmd == "\\r" {
			race, err = svc.GetRace(race.ID)
			if err != nil {
				c.ui.Error(err.Error())
				continue
			}
			if race.IsStarted() {
				c.ui.Output(fmt.Sprintf("%s (started at %s)", race.Name, race.StartTimeStr()))
			} else {
				c.ui.Output(fmt.Sprintf("%s (not started yet)", race.Name))
			}
		} else if cmd == "\\t" {
			teams, err := svc.ListTeams(race.ID)
			if err != nil {
				c.ui.Error(err.Error())
				continue
			}
			for _, team := range teams {
				filler := strings.Repeat(" ", 40-len(team.Name))
				laps := strings.Repeat("🏁  ", len(team.Laps))
				c.ui.Output(fmt.Sprintf("%d %s %s %s", team.BibNumber, team.Name, filler, laps))
			}
		} else if strings.HasPrefix(cmd, "\\d") {
			bibnumber, err := strconv.Atoi(cmd[len("\\d "):])
			if err != nil {
				c.ui.Error(err.Error())
				continue
			}
			team, lap, err := svc.DeleteLastLap(race.ID, bibnumber)
			if err != nil {
				c.ui.Error(err.Error())
				continue
			}
			c.ui.Info(fmt.Sprintf("deleted lap #%d of team %d (%s) recorded at %s", len(team.Laps)+1, team.ID, team.Name, lap.Time.Format("15:04:05")))
		} else if bibnumber, err := strconv.Atoi(cmd); err == nil {
			team, err := svc.AddLap(race.ID, bibnumber)
			if err != nil {
				c.ui.Error(err.Error())
				continue
			}
			laps := strings.Repeat("🏁  ", len(team.Laps))
			totalTime := team.Laps[len(team.Laps)-1].Time.Sub(race.StartTime).Truncate(time.Second).String()
			c.ui.Output(fmt.Sprintf("\t%s\t%s\t%s", team.Name, laps, totalTime))
		} else {
			c.ui.Info(fmt.Sprintf("unknown command: '%s'", cmd))
		}
	}

	if f, err := os.Create(history); err != nil {
		c.ui.Error(fmt.Sprintf("Error opening/creating history file: %s", err))
	} else {
		l.WriteHistory(f)
		f.Close()
	}

	c.ui.Output("bye 👋")

	return 0
}

// Synopsis return the synopsis of this command
func (c *ShellCommand) Synopsis() string {
	return "Opens the shell"
}

// Help return the help of this command
func (c *ShellCommand) Help() string {
	return `
	Usage: stopwatch shell
	  Opens the shell for interactive race management
	`
}

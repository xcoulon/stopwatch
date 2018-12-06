package cmd

import (
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/pkg/errors"

	"github.com/vatriathlon/stopwatch/service"

	"github.com/chzyer/readline"
	log "github.com/sirupsen/logrus"
)

var cmds []command

type command struct {
	name  string
	usage string
}

// Usage displays the list of available commands
func Usage(w io.Writer) {
	// nolint
	io.WriteString(w, "Usage:\n")
	for _, c := range cmds {
		_, err := io.WriteString(w, fmt.Sprintf("  %s%s%s\n", c.name, "\t", c.usage))
		if err != nil {
			log.Errorf("failed to print usage: %v", err)
			return
		}
	}

}

func init() {
	root := g.rules[0]
	log.Info("initializing help")
	cmds = []command{}
	rules := make(map[string]*rule)
	for _, r := range g.rules {
		rules[r.name] = r
	}
	if expr, ok := root.expr.(*choiceExpr); ok {
		for _, alt := range expr.alternatives {
			if alt, ok := alt.(*ruleRefExpr); ok {
				if match, ok := rules[alt.name]; ok {
					cmds = append(cmds, command{
						name:  match.name,
						usage: strings.Trim(match.displayName, `"`),
					})
				}
			}
		}
	} else {
		log.Warnf("root expression in the grammar is not a choice expression. Unable to generate help :/")
	}
}

// Command the interface for all commands
type Command interface {
	Execute(svc *service.ApplicationService, l *readline.Instance) error
}

// Execute executes the command from the given input line
func Execute(line string, svc *service.ApplicationService, l *readline.Instance) {
	if strings.TrimSpace(line) == "" {
		return
	}
	c, err := Parse("cmd", []byte(line))
	if err != nil {
		log.Errorf("unable to parse command")
		Usage(l.Stderr())
		return
	}
	log.Debugf("you said: %T", c)
	if c, ok := c.(Command); ok {
		err := c.Execute(svc, l)
		if err != nil {
			log.Errorf("failed to execute command: %v", err)
		}
	} else {
		log.Errorf("'%T' is not a valid command", c)
	}

}

// BaseCmd the base command
type BaseCmd struct {
	Name string
}

// ------------------------------------------------
// Help
// ------------------------------------------------

// HelpCmd the command to exit the program
type HelpCmd struct {
	BaseCmd
}

// NewHelpCmd returns a new HelpCmd
func NewHelpCmd() (HelpCmd, error) {
	return HelpCmd{
		BaseCmd: BaseCmd{
			Name: "help",
		},
	}, nil
}

// Execute execute the `help` command
func (c HelpCmd) Execute(svc *service.ApplicationService, l *readline.Instance) error {
	Usage(l.Stderr())
	return nil
}

// ------------------------------------------------
// List races
// ------------------------------------------------

// ListRacesCmd the command to list races
type ListRacesCmd struct {
	BaseCmd
}

// NewListRacesCmd returns a new ListRaceCmd
func NewListRacesCmd() (ListRacesCmd, error) {
	return ListRacesCmd{
		BaseCmd: BaseCmd{
			Name: "races",
		},
	}, nil
}

// Execute execute the `list races` command
func (c ListRacesCmd) Execute(svc *service.ApplicationService, l *readline.Instance) error {
	races, err := svc.ListRaces()
	if err != nil {
		return errors.Wrapf(err, "unable to list races")
	}
	log.Infof("Races:\n")
	for _, r := range races {
		log.Infof("  - %s (started: %t)\n", r.Name, r.IsStarted())
	}
	return nil
}

// ------------------------------------------------
// List teams
// ------------------------------------------------

// ListTeamsCmd the command to list races
type ListTeamsCmd struct {
	BaseCmd
}

// NewListTeamsCmd returns a new ListRaceCmd
func NewListTeamsCmd() (ListTeamsCmd, error) {
	return ListTeamsCmd{
		BaseCmd: BaseCmd{
			Name: "teams",
		},
	}, nil
}

// Execute execute the `list races` command
func (c ListTeamsCmd) Execute(svc *service.ApplicationService, l *readline.Instance) error {
	teams, err := svc.ListTeams()
	if err != nil {
		return errors.Wrapf(err, "unable to list teams")
	}
	log.Infof("Teams:\n")
	for _, t := range teams {
		log.Infof("%s - %s\n", t.BibNumber, t.Name)
	}
	return nil
}

// ------------------------------------------------
// Exit
// ------------------------------------------------

// ExitCmd the command to exit the program
type ExitCmd struct {
	BaseCmd
}

// NewExitCmd returns a new ExitCmd
func NewExitCmd() (ExitCmd, error) {
	return ExitCmd{
		BaseCmd: BaseCmd{
			Name: "exit",
		},
	}, nil
}

// Execute execute the `exit` command
func (c ExitCmd) Execute(svc *service.ApplicationService, l *readline.Instance) error {
	// log.Error("need a prompt to confirm! (need to 'stop' first?)") //TODO
	runtime.Goexit()
	return nil
}

// ------------------------------------------------
// Use race
// ------------------------------------------------

// UseRaceCmd the command to exit the program
type UseRaceCmd struct {
	BaseCmd
	RaceName string
}

// NewUseRaceCmd returns a new UseRaceCmd
func NewUseRaceCmd(raceName string) (UseRaceCmd, error) {
	return UseRaceCmd{
		BaseCmd: BaseCmd{
			Name: "race",
		},
		RaceName: strings.TrimSpace(raceName),
	}, nil
}

// Execute execute the `use race` command
func (c UseRaceCmd) Execute(svc *service.ApplicationService, l *readline.Instance) error {
	r, err := svc.UseRace(c.RaceName)
	if err != nil {
		log.Errorf("failed to use race '%s': %v", c.RaceName, err)
		return err
	}
	log.Infof("using race '%s' (started at: %s / ended at: %s)", r.Name, r.StartTimeStr(), r.EndTimeStr())
	l.SetPrompt(fmt.Sprintf("%s\033[31m»\033[0m ", r.Name))
	return nil
}

// ------------------------------------------------
// Start race
// ------------------------------------------------

// StartRaceCmd the command to start the current race
type StartRaceCmd struct {
	BaseCmd
}

// NewStartRaceCmd returns a new StartRaceCmd
func NewStartRaceCmd() (StartRaceCmd, error) {
	return StartRaceCmd{
		BaseCmd: BaseCmd{
			Name: "start",
		},
	}, nil
}

// Execute execute the "start" command
func (c StartRaceCmd) Execute(svc *service.ApplicationService, l *readline.Instance) error {
	startedAt, err := svc.StartCurrentRace()
	if err != nil {
		return errors.Wrap(err, "failed to start race")
	}
	log.Infof("race started at %v", startedAt)
	return nil
}

// ------------------------------------------------
// Stop race
// ------------------------------------------------

// StopRaceCmd the command to end the current race
type StopRaceCmd struct {
	BaseCmd
}

// NewStopRaceCmd returns a new StartRaceCmd
func NewStopRaceCmd() (StopRaceCmd, error) {
	return StopRaceCmd{
		BaseCmd: BaseCmd{
			Name: "stop",
		},
	}, nil
}

// Execute execute the "stop" command
func (c StopRaceCmd) Execute(svc *service.ApplicationService, l *readline.Instance) error {
	// e, err := svc.StopRace(context.Background())
	// if err != nil {
	// 	log.Error("failed to end race '%s': %v", c.RaceName, err)
	// }
	log.Error("not implemented yet!")
	return nil
}

// ------------------------------------------------
// Add lap
// ------------------------------------------------

// AddLapCmd the command to exit the program
type AddLapCmd struct {
	BaseCmd
	bibnumber string
}

// NewAddLapCmd returns a new AddLapCmd
func NewAddLapCmd(bibnumber string) (AddLapCmd, error) {
	return AddLapCmd{
		BaseCmd: BaseCmd{
			Name: "lap",
		},
		bibnumber: bibnumber,
	}, nil
}

// Execute execute the `help` command
func (c AddLapCmd) Execute(svc *service.ApplicationService, l *readline.Instance) error {
	log.Infof("adding 1 lap to team #%s", c.bibnumber)
	return nil
}

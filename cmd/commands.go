package cmd

import (
	"context"
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
		log.Errorf("unable to parse command: %v", err)
		Usage(l.Stderr())
	}
	// log.Debugf("you said: %T", c)
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
			Name: "list",
		},
	}, nil
}

func (c ListRacesCmd) Execute(svc *service.ApplicationService, l *readline.Instance) error {
	races, err := svc.ListRaces(context.Background())
	if err != nil {
		return errors.Wrapf(err, "unable to list races")
	}
	fmt.Fprintf(l.Stdout(), "Races:\n")
	for _, r := range races {
		fmt.Fprintf(l.Stdout(), "  - %s (started: %t)\n", r.Name, r.IsStarted())
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

func (c ExitCmd) Execute(svc *service.ApplicationService, l *readline.Instance) error {
	log.Error("need a prompt to confirm! (need to 'stop' first?)") //TODO
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
			Name: "use",
		},
		RaceName: strings.TrimSpace(raceName),
	}, nil
}

func (c UseRaceCmd) Execute(svc *service.ApplicationService, l *readline.Instance) error {
	r, err := svc.UseRace(context.Background(), c.RaceName)
	if err != nil {
		fmt.Fprintf(l.Stderr(), "failed to use race '%s': %v", c.RaceName, err)
	}
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
	err := svc.StartCurrentRace(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to start race")
	}
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
	// 	fmt.Fprintf(l.Stderr(), "failed to end race '%s': %v", c.RaceName, err)
	// }
	return nil
}

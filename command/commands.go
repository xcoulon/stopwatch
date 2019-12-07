package command

import (
	"fmt"
	"os"

	"github.com/mattn/go-colorable"
	"github.com/mitchellh/cli"
)

// Run runs the command with the given args
func Run(args []string) int {
	ui := &cli.ColoredUi{
		ErrorColor: cli.UiColorRed,
		WarnColor:  cli.UiColorYellow,
		InfoColor:  cli.UiColorGreen,
		Ui: &cli.BasicUi{
			Reader:      os.Stdin,
			Writer:      colorable.NewColorableStdout(),
			ErrorWriter: colorable.NewColorableStderr(),
		},
	}

	// Commands is the mapping of all the available commands.
	commands := map[string]cli.CommandFactory{
		"import-teams": func() (cli.Command, error) {
			return NewImportTeamsCommand(ui), nil
		},
		"export-results": func() (cli.Command, error) {
			return NewGenerateResultsCommand(ui), nil
		},
		"list-races": func() (cli.Command, error) {
			return NewListRacesCommand(ui), nil
		},
		"shell": func() (cli.Command, error) {
			return NewShellCommand(ui), nil
		},
	}

	cli := &cli.CLI{
		Name:                       "stopwatch",
		Args:                       args,
		Commands:                   commands,
		Autocomplete:               true,
		AutocompleteNoDefaultFlags: true,
	}

	exitCode, err := cli.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing CLI: %s\n", err.Error())
		return 1
	}

	return exitCode

}

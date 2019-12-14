package command

import (
	"flag"
	"fmt"

	"github.com/mitchellh/cli"
	"github.com/vatriathlon/stopwatch/pkg/configuration"
	"github.com/vatriathlon/stopwatch/pkg/connection"
	"github.com/vatriathlon/stopwatch/pkg/service"
)

// ImportTeamsCommand the command to import teams in the database
type ImportTeamsCommand struct {
	ui    cli.Ui
	flags *flag.FlagSet
	help  string

	file string
}

var _ cli.Command = (*ImportTeamsCommand)(nil)

// var _ cli.CommandAutocomplete = (*ImportTeamsCommand)(nil)

// NewImportTeamsCommand returns a new command to import teams in the database
func NewImportTeamsCommand(ui cli.Ui) *ImportTeamsCommand {
	c := &ImportTeamsCommand{ui: ui}
	c.init()
	return c
}

func (c *ImportTeamsCommand) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.file, "file", "", "The source file to import.")
}

// Run executes the command
func (c *ImportTeamsCommand) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	if c.file == "" {
		c.ui.Error("the --file parameter must be specified")
		return 1
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
	svc := service.NewImportService(db)
	err = svc.ImportFromFile(c.file)
	if err != nil {
		c.ui.Error(fmt.Sprintf("error while importing from '%s': %s", c.file, err.Error()))
		return 1
	}

	return 0
}

// Synopsis return the synopsis of this command
func (c *ImportTeamsCommand) Synopsis() string {
	return "Import teams in the database"
}

// Help return the help of this command
func (c *ImportTeamsCommand) Help() string {
	return `
	Usage: stopwatch import --file=<path/to/file.csv>
	  Imports the teams specified by the CSV file in the database
	`
}

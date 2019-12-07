package command

import (
	"fmt"

	"github.com/mitchellh/cli"
	"github.com/vatriathlon/stopwatch/pkg/configuration"
	"github.com/vatriathlon/stopwatch/pkg/connection"
	"github.com/vatriathlon/stopwatch/pkg/service"
)

// ListRacesCommand the command to import teams in the database
type ListRacesCommand struct {
	ui   cli.Ui
	help string
}

var _ cli.Command = (*ListRacesCommand)(nil)

// var _ cli.CommandAutocomplete = (*ListRacesCommand)(nil)

// NewListRacesCommand returns a new command to import teams in the database
func NewListRacesCommand(ui cli.Ui) *ListRacesCommand {
	c := &ListRacesCommand{
		ui: ui,
	}
	return c
}

// Run executes the command
func (c *ListRacesCommand) Run(args []string) int {
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
	races, err := svc.ListRaces()
	if err != nil {
		c.ui.Error(fmt.Sprintf("error while listing races: %s", err.Error()))
		return 1
	}
	c.ui.Info("ID\tName")
	for _, r := range races {
		c.ui.Info(fmt.Sprintf("%d\t%s", r.ID, r.Name))

	}
	return 0
}

// Synopsis return the synopsis of this command
func (c *ListRacesCommand) Synopsis() string {
	return "Lists the races"
}

// Help return the help of this command
func (c *ListRacesCommand) Help() string {
	return `
	Usage: \?
	  Lists the races
	`
}

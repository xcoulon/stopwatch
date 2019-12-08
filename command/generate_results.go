package command

import (
	"flag"
	"fmt"
	"os/exec"

	"github.com/mitchellh/cli"
	"github.com/vatriathlon/stopwatch/pkg/configuration"
	"github.com/vatriathlon/stopwatch/pkg/connection"
	"github.com/vatriathlon/stopwatch/pkg/service"
)

// GenerateResultsCommand the command to import teams in the database
type GenerateResultsCommand struct {
	ui        cli.Ui
	flags     *flag.FlagSet
	help      string
	raceID    int
	outputDir string
}

var _ cli.Command = (*GenerateResultsCommand)(nil)

// NewGenerateResultsCommand returns a new command to generate the results
func NewGenerateResultsCommand(ui cli.Ui) *GenerateResultsCommand {
	c := &GenerateResultsCommand{ui: ui}
	c.init()
	return c
}

func (c *GenerateResultsCommand) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.IntVar(&c.raceID, "race", 0, "The ID of the race")
	c.flags.StringVar(&c.outputDir, "output", "", "The target directory for the reports")
}

// Run executes the command
func (c *GenerateResultsCommand) Run(args []string) int {
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	if c.outputDir == "" {
		c.ui.Error("the --output parameter must be specified")
		return 1
	}
	if c.raceID == 0 {
		c.ui.Error("the --race parameter must be specified")
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
	svc := service.NewResultService(db)
	files, err := svc.GenerateResults(c.raceID, c.outputDir)
	if err != nil {
		c.ui.Error(fmt.Sprintf("error while generating the reports for race '%d': %s", c.raceID, err.Error()))
		return 1
	}
	for _, file := range files {
		// external call to asciidoctor-pdf
		c.ui.Output(fmt.Sprintf("generating pdf for %s", file))
		err = exec.Command("asciidoctor-pdf", file).Run()
		if err != nil {
			c.ui.Error(err.Error())
		}
	}

	return 0
}

// Synopsis return the synopsis of this command
func (c *GenerateResultsCommand) Synopsis() string {
	return "Generates the results for the given race"
}

// Help return the help of this command
func (c *GenerateResultsCommand) Help() string {
	return `Usage: stopwatch export --race=<id> --output=<path/to/dir>
	Generates the results for the given race and exports in the given output directory
`
}

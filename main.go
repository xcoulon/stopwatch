package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/vatriathlon/stopwatch/configuration"
	"github.com/vatriathlon/stopwatch/connection"
	"github.com/vatriathlon/stopwatch/server"
	"github.com/vatriathlon/stopwatch/service"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq" // need to import postgres driver
	"github.com/sirupsen/logrus"
)

// initializes the level for the logger, using the optional '-debug' flag to activate the logs in 'debug' level.
// Other tests must import this 'test' package even if unused, using:
// import _ "github.com/bytesparadise/libasciidoc/pkg/log"
func init() {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.DisableLevelTruncation = true
	customFormatter.DisableTimestamp = true
	customFormatter.ForceColors = true
	logrus.SetFormatter(customFormatter)
}

func main() {
	var importFile string
	var outputFile string
	var generateResults bool
	var raceID int
	flag.StringVar(&importFile, "import", "", "imports the file in the database.")
	flag.BoolVar(&generateResults, "result", false, "flag to genefrate the race results")
	flag.IntVar(&raceID, "raceID", 0, "id of the race for the results")
	flag.StringVar(&outputFile, "output", "", "id of the race for the results")
	flag.Parse()

	config, err := configuration.New()
	if err != nil {
		panic(err)
	}
	db, err := connection.NewUserConnection(config)
	if err != nil {
		logrus.Fatalf("failed to start: %s", err.Error())
	}

	db.LogMode(config.IsDBLogsEnabled())

	// handle shutdown
	go handleShutdown(db)

	if importFile != "" {
		logrus.WithField("file", importFile).Info("importing...")
		svc := service.NewImportService(db)
		err := svc.ImportFromFile(importFile)
		if err != nil {
			logrus.Fatalf("failed to import from file: %s", err.Error())
		}
		return
	}

	if generateResults {
		logrus.WithField("race_id", raceID).WithField("output_file", outputFile).Info("Generating results...")
		svc := service.NewResultService(db)
		err := svc.GenerateResults(raceID, outputFile)
		if err != nil {
			logrus.Fatalf("failed to export result: %s", err.Error())
		}
		return

	}

	s := server.New(service.NewApplicationService(db))
	// listen and serve on 0.0.0.0:8080
	s.Start(":8080")
}

func handleShutdown(db *gorm.DB) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	// handle ctrl+c event here
	// for example, close database
	logrus.Warn("Closing DB connection before complete shutdown")
	err := db.Close()
	if err != nil {
		logrus.Errorf("error while closing the connection to the database: %v", err)
	}
	os.Exit(0)
}

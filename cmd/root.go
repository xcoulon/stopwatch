package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/vatriathlon/stopwatch/pkg/configuration"
	"github.com/vatriathlon/stopwatch/pkg/connection"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq" // need to import postgres driver
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.DisableLevelTruncation = true
	customFormatter.DisableTimestamp = true
	customFormatter.ForceColors = true
	logrus.SetFormatter(customFormatter)
}

var db *gorm.DB

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{}

func openDB() {
	config, err := configuration.New()
	if err != nil {
		panic(err)
	}
	db, err = connection.New(config)
	if err != nil {
		logrus.Fatalf("failed to start: %s", err.Error())
	}

	db.LogMode(config.IsDBLogsEnabled())
}

func closeDB() {
	logrus.Warn("Closing DB connection before complete shutdown")
	err := db.Close()
	if err != nil {
		logrus.Errorf("error while closing the connection to the database: %v", err)
	}
	os.Exit(0)
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

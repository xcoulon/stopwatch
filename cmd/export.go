package cmd

import (
	"github.com/vatriathlon/stopwatch/pkg/service"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	exportCmd.Flags().IntVar(&raceID, "raceID", 0, "id of the race for the results")
	exportCmd.MarkFlagRequired("raceID")
	exportCmd.Flags().StringVar(&outputFile, "output", "", "id of the race for the results")
	exportCmd.MarkFlagRequired("output")
	rootCmd.AddCommand(exportCmd)
}

var raceID int
var outputFile string

// exportCmd represents the result export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "exports the results",
	PreRun: func(cmd *cobra.Command, args []string) {
		openDB()
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		closeDB()
	},
	Run: func(cmd *cobra.Command, args []string) {
		logrus.WithField("race_id", raceID).WithField("output_file", outputFile).Info("Generating results...")
		svc := service.NewResultService(db)
		err := svc.GenerateResults(raceID, outputFile)
		if err != nil {
			logrus.Fatalf("failed to export result: %s", err.Error())
		}
		return

	},
}

package cmd

import (
	"github.com/vatriathlon/stopwatch/pkg/service"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	importCmd.Flags().StringVar(&importFile, "file", "", "imports the file in the database.")
	importCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(importCmd)
}

var importFile string

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "imports the CSV data in the DB",
	PreRun: func(cmd *cobra.Command, args []string) {
		openDB()
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		closeDB()
	},
	Run: func(cmd *cobra.Command, args []string) {
		logrus.WithField("file", importFile).Info("importing...")
		svc := service.NewImportService(db)
		err := svc.ImportFromFile(importFile)
		if err != nil {
			logrus.Fatalf("failed to import from file: %s", err.Error())
		}
		return
	},
}

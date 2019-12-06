package cmd

import (
	"github.com/vatriathlon/stopwatch/pkg/server"
	"github.com/vatriathlon/stopwatch/pkg/service"
	
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "runs the server",
	PreRun: func(cmd *cobra.Command, args []string) {
		openDB()
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		closeDB()
	},
	Run: func(cmd *cobra.Command, args []string) { 
		s := server.New(service.NewApplicationService(db))
		// listen and serve on 0.0.0.0:8080
		s.Start(":8080")
	},
}

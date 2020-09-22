package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of kubectl node-pod",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("kubectl node-pod version 0.0.1")
	},
}

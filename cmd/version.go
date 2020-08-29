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
	Short: "Print the version number of kube-node-pod",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("kube-node-pod version 0.0.1")
	},
}

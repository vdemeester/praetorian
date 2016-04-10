package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vdemeester/praetorian/version"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display praetorian version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version, version.GitCommit)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

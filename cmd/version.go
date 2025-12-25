package cmd

import (
	"github.com/josexy/catgo/version"
	"github.com/spf13/cobra"
)

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version information.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		version.Show()
		return nil
	},
}

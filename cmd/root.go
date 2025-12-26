package cmd

import (
	"os"

	"github.com/josexy/catgo/internal/util"
	"github.com/josexy/catgo/version"
	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use:           "catgo",
	Short:         "Simple Go's package manager like Cargo",
	Version:       version.Version,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		util.CheckGoInstalled()
	},
}

func init() {
	rootCommand.AddCommand(runCommand)
	rootCommand.AddCommand(buildCommand)
	rootCommand.AddCommand(cleanCommand)
	rootCommand.AddCommand(newCommand)
	rootCommand.AddCommand(initCommand)
	rootCommand.AddCommand(addCommand)
	rootCommand.AddCommand(removeCommand)
	rootCommand.AddCommand(versionCommand)
	rootCommand.AddCommand(vendorCommand)
	rootCommand.AddCommand(testCommand)
}

func Execute() {
	if err := rootCommand.Execute(); err != nil {
		util.Printer.PrintError(err.Error())
		os.Exit(1)
	}
}

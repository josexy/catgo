package cmd

import (
	"fmt"

	"github.com/josexy/catgo/internal/util"
	"github.com/spf13/cobra"
)

var newCommand = &cobra.Command{
	Use:   "new [OPTIONS] <path>",
	Short: "Create a new package",
	Long: `Create a new Go package at the specified path.

  This command will create a new directory and initialize the Go module.`,
	RunE: runNew,
}

func init() {
	newCommand.Flags().StringVar(&initName, "name", "", "Set the resulting package name")
	newCommand.Flags().BoolVar(&initGitRepo, "git", false, "Initialize a new repository with Git")
}

func runNew(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("path is required")
	}
	projectDir := args[0]
	if projectDir != "." {
		if err := util.Mkdir(projectDir); err != nil {
			return err
		}
	}

	return runInit(cmd, []string{projectDir})
}

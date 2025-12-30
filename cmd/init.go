package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/josexy/catgo/internal/util"
	"github.com/josexy/catgo/internal/util/template"
	"github.com/spf13/cobra"
)

var (
	initName    string
	initGitRepo bool
)

var initCommand = &cobra.Command{
	Use:   "init [OPTIONS] [path]",
	Short: "Create a new package in an existing directory",
	Long: `Create a new Go package in an existing directory.

  This command will initialize a Go module in the current or specified directory.`,
	RunE: runInit,
}

func init() {
	initCommand.Flags().StringVar(&initName, "name", "", "Set the resulting package name")
	initCommand.Flags().BoolVar(&initGitRepo, "git", false, "Initialize a new repository with Git")
}

func runInit(cmd *cobra.Command, args []string) error {
	projectDir := "."
	if len(args) > 0 {
		projectDir = args[0]
	}

	if projectDir != "." {
		if err := util.Mkdir(projectDir); err != nil {
			return err
		}
		if err := os.Chdir(projectDir); err != nil {
			return fmt.Errorf("could not change to directory: %w", err)
		}
	}

	if util.PathExist("go.mod") {
		return fmt.Errorf("go.mod already exists")
	}

	moduleName := initName
	if moduleName == "" {
		cwd, err := util.CurrentDir()
		if err != nil {
			return err
		}
		moduleName = filepath.Base(cwd)
	}

	util.Printer.PrintCreated(fmt.Sprintf("package `%s`", moduleName))

	if err := util.Exec(context.Background(), "go", []string{"mod", "init", moduleName}, nil); err != nil {
		return err
	}

	if !util.PathExist("main.go") {
		if err := util.WriteFile("main.go", []byte(template.GoMainFile)); err != nil {
			return err
		}
	}

	if !util.PathExist(".gitignore") {
		if err := util.WriteFile(".gitignore", []byte(template.GitIgnoreFile)); err != nil {
			return err
		}
	}

	if initGitRepo {
		if !util.PathExist(".git") {
			if err := util.Exec(context.Background(), "git", []string{"init"}, nil); err != nil {
				util.Printer.PrintWarning(fmt.Sprintf("could not initialize git repository: %v", err))
			}
		}
	}

	return nil
}

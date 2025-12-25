package cmd

import (
	"fmt"
	"strings"

	"github.com/josexy/catgo/internal/util"
	"github.com/spf13/cobra"
)

var (
	dependencyVersion string
)

var addCommand = &cobra.Command{
	Use:   "add [OPTIONS] <package>...",
	Short: "Add dependencies to a Go project",
	Long: `Add one or more dependencies to the project.

  This command will add the specified packages to go.mod via go get.

  If there is only one dependency, the revision can be specified with --rev.

  However, if there is more than one dependency, the revision will be ignored.`,
	RunE: runAdd,
}

func init() {
	addCommand.Flags().StringVar(&dependencyVersion, "rev", "", "Specific commit to use when adding from git")
}

func runAdd(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no dependencies specified")
	}

	moduleName, err := util.CurrentModuleName()
	if err != nil {
		return err
	}

	for _, pkg := range args {
		pkg = strings.TrimSpace(pkg)
		if pkg == "" {
			continue
		}

		// only add git references if there is only one dependency
		if len(args) == 1 && dependencyVersion != "" && !strings.Contains(pkg, "@") {
			pkg = fmt.Sprintf("%s@%s", pkg, dependencyVersion)
		}

		util.Printer.PrintUpdating(fmt.Sprintf("module %s go.mod and go.sum", moduleName))
		util.Printer.PrintAdding(pkg)

		if err = util.Exec("go", []string{"get", pkg}, nil); err != nil {
			return err
		}
	}

	util.Printer.PrintWarning("Please run `go mod tidy` to update go.mod and go.sum")

	return nil
}

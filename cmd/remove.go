package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/josexy/catgo/internal/util"
	"github.com/spf13/cobra"
)

var removeCommand = &cobra.Command{
	Use:   "remove [OPTIONS] <package>...",
	Short: "Remove dependencies from a Go project",
	Long: `Remove one or more dependencies from the project.

  This command will remove the specified packages from go.mod and run go mod tidy.`,
	RunE: runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no dependencies specified")
	}

	goModPath, err := util.CurrentGoModFile()
	if err != nil {
		return err
	}

	file, err := os.Open(goModPath)
	if err != nil {
		return fmt.Errorf("could not open go.mod: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("could not read go.mod: %w", err)
	}

	for _, pkg := range args {
		pkg = strings.TrimSpace(pkg)
		if pkg == "" {
			continue
		}

		util.Printer.PrintRemoving(pkg)

		var newLines []string
		removed := false
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			if strings.HasPrefix(trimmedLine, pkg+" ") {
				removed = true
				continue
			}
			newLines = append(newLines, line)
		}

		if removed {
			lines = newLines
		} else {
			util.Printer.PrintWarning(fmt.Sprintf("package %s not found in go.mod", pkg))
		}
	}

	content := strings.Join(lines, "\n")
	if err = util.WriteFile(goModPath, []byte(content)); err != nil {
		return err
	}

	util.Printer.PrintUpdating("go.mod and go.sum")
	if err = util.Exec(context.Background(), "go", []string{"mod", "tidy"}, nil); err != nil {
		return err
	}
	return nil
}

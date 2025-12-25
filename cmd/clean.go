package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/josexy/catgo/internal/util"
	"github.com/spf13/cobra"
)

var cleanCommand = &cobra.Command{
	Use:   "clean [OPTIONS]",
	Short: "Remove all generated binaries for the local package",
	Long:  `Remove all generated binaries for the local package.`,
	RunE:  runClean,
}

func runClean(cmd *cobra.Command, args []string) error {
	goModDir, err := util.CurrentGoModDir()
	if err != nil {
		return err
	}
	moduleName, err := util.CurrentModuleName()
	if err != nil {
		return err
	}
	parts := strings.Split(moduleName, "/")
	target := filepath.Join(goModDir, "bin", parts[len(parts)-1])

	var removed []string
	patterns := []string{target, target + "-*"}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			if err := os.Remove(match); err == nil {
				removed = append(removed, match)
			}
		}
	}

	util.Printer.PrintRemoved(removed)
	return nil
}

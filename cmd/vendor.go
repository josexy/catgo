package cmd

import (
	"fmt"

	"github.com/josexy/catgo/internal/util"
	"github.com/spf13/cobra"
)

var vendorCommand = &cobra.Command{
	Use:   "vendor",
	Short: "Vendor all dependencies for a project locally",
	Long:  `Vendor all dependencies for a project locally to the vendor directory.`,
	RunE:  runVendor,
}

func runVendor(cmd *cobra.Command, args []string) error {
	moduleName, err := util.CurrentModuleName()
	if err != nil {
		return err
	}

	util.Printer.PrintVendoring(fmt.Sprintf("vendor %s dependencies", moduleName))

	if err := util.Exec("go", []string{"mod", "vendor"}, nil); err != nil {
		return err
	}

	return nil
}

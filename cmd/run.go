package cmd

import (
	"path/filepath"

	"github.com/josexy/catgo/internal/util"
	"github.com/spf13/cobra"
)

var runCommand = &cobra.Command{
	Use:   "run [OPTIONS] [-- ARGS]",
	Short: "Compile and run a binary of the local package",
	Long: `Compile and run a binary of the local package.

  All the arguments following the two dashes (--) are passed to the binary
  to run. If you're passing arguments to both Catgo and the binary, the
  ones after -- go to the binary, the ones before go to Catgo.

  By default, Catgo builds the binary in the bin/ directory of the current
  package.

  This Catgo uses the current Go module to build(via "go env GOMOD"). And 
  you can specify the package to build with the --package flag.`,
	RunE: runRun,
}

func init() {
	runCommand.Flags().StringVarP(&buildTarget, "target", "t", "", "Build for the target triple, e.g. linux/amd64")
	runCommand.Flags().BoolVarP(&buildRelease, "release", "r", false, "Build artifacts in release mode, with optimizations")
	runCommand.Flags().StringVarP(&buildOutput, "output", "o", "", "Output binary name")
	runCommand.Flags().StringVarP(&buildPackage, "package", "p", "", "Package to build")
	runCommand.Flags().BoolVarP(&buildLocal, "local", "l", false, "Build to current directory, default to bin/ directory")
	runCommand.Flags().BoolVarP(&buildCGOZero, "cgo-zero", "z", false, "Build with CGO disabled")
	runCommand.Flags().BoolVar(&buildVendor, "vendor", false, "Build with vendor directory, if a vendor directory exists it will be used")
	runCommand.Flags().StringSliceVarP(&buildSetVariables, "set", "x", nil, "Set Go build flags -X")
}

func runRun(cmd *cobra.Command, args []string) error {
	target, err := runBuild(cmd, args)
	if err != nil {
		return err
	}

	goModDir, err := util.CurrentGoModDir()
	if err != nil {
		return err
	}
	relTarget, _ := filepath.Rel(goModDir, target)
	if relTarget == "" {
		relTarget = target
	}

	util.Printer.PrintRunning(util.FormatCommandArgs(relTarget, args))
	if err = util.ExecProcess(target, args, nil); err != nil {
		return err
	}
	return nil
}

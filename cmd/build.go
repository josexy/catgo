package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/josexy/catgo/internal/util"
	"github.com/spf13/cobra"
)

var (
	buildRelease      bool
	buildOutput       string
	buildPackage      string
	buildTarget       string
	buildLocal        bool
	buildCGOZero      bool
	buildVendor       bool
	buildSetVariables []string
)

var buildCommand = &cobra.Command{
	Use:   "build [OPTIONS]",
	Short: "Compile the local package to a binary",
	Long: `Compile the local package to a binary.

  This command will compile the local package to a binary and place it in the bin directory. 

  The binary name will be the package name if not specified.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := runBuild(cmd, args)
		return err
	},
}

func init() {
	buildCommand.Flags().StringVarP(&buildTarget, "target", "t", "", "Build for the target triple, e.g. linux/amd64")
	buildCommand.Flags().BoolVarP(&buildRelease, "release", "r", false, "Build artifacts in release mode, with optimizations")
	buildCommand.Flags().StringVarP(&buildOutput, "output", "o", "", "Output binary name, default to package name")
	buildCommand.Flags().StringVarP(&buildPackage, "package", "p", "", "Package to build")
	buildCommand.Flags().BoolVarP(&buildLocal, "local", "l", false, "Build to current directory, default to bin directory")
	buildCommand.Flags().BoolVarP(&buildCGOZero, "cgo-zero", "z", false, "Build with CGO disabled")
	buildCommand.Flags().BoolVar(&buildVendor, "vendor", false, "Build with vendor directory, if a vendor directory exists it will be used")
	buildCommand.Flags().StringSliceVarP(&buildSetVariables, "set", "x", nil, "Set Go build flags -X")
}

func runBuild(_ *cobra.Command, _ []string) (string, error) {
	startTime := time.Now()

	moduleName, err := util.CurrentModuleName()
	if err != nil {
		return "", err
	}

	target := buildOutput
	if target == "" {
		parts := strings.Split(moduleName, "/")
		target = parts[len(parts)-1]
	} else {
		target = filepath.Base(target)
	}

	var env []string

	targetOS, targetArch, target := parseBuildTarget(target, buildTarget)
	if targetOS != "" {
		env = append(env, fmt.Sprintf("GOOS=%s", targetOS))
	}
	if targetArch != "" {
		env = append(env, fmt.Sprintf("GOARCH=%s", targetArch))
	}

	if buildLocal {
		currentDir, err := util.CurrentDir()
		if err != nil {
			return "", err
		}
		target = filepath.Join(currentDir, target)
	} else {
		goModDir, err := util.CurrentGoModDir()
		if err != nil {
			return "", err
		}
		outputDir := filepath.Join(goModDir, "bin")
		if err = util.Mkdir(outputDir); err != nil {
			return "", err
		}
		target = filepath.Join(outputDir, target)
	}

	util.Printer.PrintCompiling(fmt.Sprintf("%s (%s)", moduleName, target))

	bldArgs := []string{"build", "-o", target}
	if buildVendor {
		bldArgs = append(bldArgs, "-mod=vendor")
	}
	if buildRelease {
		bldArgs = append(bldArgs, "-trimpath")
	}
	if buildRelease || len(buildSetVariables) > 0 {
		bldArgs = append(bldArgs, "-ldflags")
	}
	var ldflags []string
	if buildRelease {
		ldflags = append(ldflags, "-s", "-w")
	}
	if len(buildSetVariables) > 0 {
		for _, v := range buildSetVariables {
			ldflags = append(ldflags, fmt.Sprintf("-X '%s'", v))
		}
	}
	if len(ldflags) > 0 {
		bldArgs = append(bldArgs, strings.Join(ldflags, " "))
	}

	if buildCGOZero {
		env = append(env, "CGO_ENABLED=0")
	}

	if buildPackage, err = parseToGoPackage(moduleName, buildPackage); err != nil {
		return "", err
	}

	bldArgs = append(bldArgs, buildPackage)

	if err = util.Exec(context.Background(), "go", bldArgs, env); err != nil {
		return "", err
	}

	profile := "dev"
	if buildRelease {
		profile = "release"
	}
	util.Printer.PrintFinished(profile, util.FormatDuration(time.Since(startTime)))
	return target, nil
}

func parseToGoPackage(moduleName, packageName string) (string, error) {
	goModDir, err := util.CurrentGoModDir()
	if err != nil {
		return "", err
	}
	currentDir, err := util.CurrentDir()
	if err != nil {
		return "", err
	}
	if packageName == "" || packageName == "." {
		relDir, _ := filepath.Rel(goModDir, currentDir)
		packageName = path.Join(moduleName, relDir)
	} else if !strings.HasPrefix(packageName, moduleName) {
		info, err := os.Stat(packageName)
		if errors.Is(err, os.ErrNotExist) || info == nil {
			return "", fmt.Errorf("package `%s` does not exist", packageName)
		}
		var relDir string
		if info.IsDir() {
			relDir, _ = filepath.Rel(goModDir, filepath.Join(currentDir, packageName))
		} else {
			relDir, _ = filepath.Rel(goModDir, filepath.Join(currentDir, filepath.Dir(packageName)))
		}
		relDir = path.Join(filepath.SplitList(relDir)...)
		packageName = path.Join(moduleName, relDir)
	}
	return packageName, nil
}

func parseBuildTarget(name, buildTarget string) (targetOS, targetArch, targetName string) {
	const exeExt = ".exe"
	var needWindowsExeExt bool
	if buildTarget != "" {
		parts := strings.Split(buildTarget, "/")
		if len(parts) > 0 {
			targetOS = parts[0]
			if strings.Contains(parts[0], "windows") && !strings.HasSuffix(name, exeExt) {
				needWindowsExeExt = true
			}
		}
		if len(parts) > 1 {
			targetArch = parts[1]
		}
	} else if runtime.GOOS == "windows" && !strings.HasSuffix(name, exeExt) {
		needWindowsExeExt = true
	}
	targetName = name
	if targetOS != "" {
		targetName += "-" + targetOS
	}
	if targetArch != "" {
		targetName += "-" + targetArch
	}
	if needWindowsExeExt {
		targetName += exeExt
	}
	return
}

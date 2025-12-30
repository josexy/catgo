package cmd

import (
	"context"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/josexy/catgo/internal/test"
	"github.com/josexy/catgo/internal/util"
	"github.com/spf13/cobra"
)

const allPackagesSuffix = "/..."

var (
	testCount    int
	testVerbose  bool
	testRace     bool
	testFullPath bool
	testFailFast bool
	testPackage  string
	testRunExpr  string
	testSkipExpr string
	testTimeout  time.Duration
)

var testCommand = &cobra.Command{
	Use:   "test [OPTIONS] [TESTNAME] [-- [ARGS]...]",
	Short: "Test the local package",
	Long:  `Test the local package.`,
	RunE:  runTest,
}

func init() {
	testCommand.Flags().StringVarP(&testRunExpr, "run", "r", "^Test", "Run only those tests matching the regular expression")
	testCommand.Flags().StringVarP(&testSkipExpr, "skip", "s", "", "Skip tests matching the regular expression")
	testCommand.Flags().StringVarP(&testPackage, "package", "p", "./...", "The package to test, default to all packages")
	testCommand.Flags().IntVarP(&testCount, "count", "c", 1, "The number of times to run each test")
	testCommand.Flags().DurationVarP(&testTimeout, "timeout", "t", 0, "The time limit for each test")
	testCommand.Flags().BoolVar(&testRace, "race", false, "Enable race detector")
	testCommand.Flags().BoolVarP(&testVerbose, "verbose", "v", false, "Show output from tests")
	testCommand.Flags().BoolVar(&testFullPath, "fullpath", false, "Show full file names in error messages")
	testCommand.Flags().BoolVar(&testFailFast, "failfast", false, "Do not start new tests after the first test failure")
}

func runTest(cmd *cobra.Command, args []string) error {
	moduleName, err := util.CurrentModuleName()
	if err != nil {
		return err
	}
	return execGoTest(moduleName, args)
}

func execGoTest(moduleName string, args []string) error {
	var testAllPackages bool
	if strings.HasSuffix(testPackage, allPackagesSuffix) {
		testPackage = strings.TrimSuffix(testPackage, allPackagesSuffix)
		testAllPackages = true
	}
	testPackage, err := parseToGoPackage(moduleName, testPackage)
	if err != nil {
		return err
	}
	if testAllPackages {
		testPackage = testPackage + allPackagesSuffix
	}

	testArgs := []string{"test"}
	if testRunExpr != "" {
		testArgs = append(testArgs, "-run", testRunExpr)
	}
	if testSkipExpr != "" {
		testArgs = append(testArgs, "-skip", testSkipExpr)
	}
	if testFailFast {
		testArgs = append(testArgs, "-failfast")
	}
	if testFullPath {
		testArgs = append(testArgs, "-fullpath")
	}
	if testCount > 0 {
		testArgs = append(testArgs, "-count", strconv.Itoa(testCount))
	}
	if testTimeout > 0 {
		testArgs = append(testArgs, "-timeout", testTimeout.String())
	}
	if testRace {
		testArgs = append(testArgs, "-race")
	}
	testArgs = append(testArgs, "-json")
	testArgs = append(testArgs, testPackage)
	testArgs = append(testArgs, args...)

	errCh := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	pr, pw := io.Pipe()
	analyzer := test.New(pr, moduleName, testVerbose)

	go func() {
		defer pw.Close()
		errCh <- util.Exec(ctx, "go", testArgs, nil, util.ExecIO{Stdout: pw, Stderr: pw})
	}()

	analyzer.Wait()
	pr.Close()
	cancel()
	return <-errCh
}

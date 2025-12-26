package cmd

import (
	"io"
	"path"
	"strconv"
	"time"

	"github.com/josexy/catgo/internal/test"
	"github.com/josexy/catgo/internal/util"
	"github.com/spf13/cobra"
)

var (
	testCount   int
	testVerbose bool
	testRace    bool
	testPackage string
	testRunF    string
	testTimeout time.Duration
)

var testCommand = &cobra.Command{
	Use:   "test [OPTIONS] [TESTNAME] [-- [ARGS]...]",
	Short: "Test the local package",
	Long:  `Test the local package.`,
	RunE:  runTest,
}

func init() {
	testCommand.Flags().StringVarP(&testRunF, "run", "r", "^Test", "Run only those tests matching the regular expression")
	testCommand.Flags().StringVarP(&testPackage, "package", "p", "", "The package to test, default to current package")
	testCommand.Flags().IntVarP(&testCount, "count", "c", 1, "The number of times to run each test")
	testCommand.Flags().DurationVarP(&testTimeout, "timeout", "t", 0, "The time limit for each test")
	testCommand.Flags().BoolVar(&testRace, "race", false, "Enable race detector")
	testCommand.Flags().BoolVarP(&testVerbose, "verbose", "v", false, "Show output from tests")
}

func runTest(cmd *cobra.Command, args []string) error {
	moduleName, err := util.CurrentModuleName()
	if err != nil {
		return err
	}
	return execGoTest(moduleName, args)
}

func execGoTest(moduleName string, args []string) error {
	testPackage, err := parseToGoPackage(moduleName, testPackage)
	if err != nil {
		return err
	}
	testPackage = path.Clean(testPackage) + "/..."

	testArgs := []string{"test"}
	if testRunF != "" {
		testArgs = append(testArgs, "-run", testRunF)
	}
	if testCount > 0 {
		testArgs = append(testArgs, "-count", strconv.Itoa(testCount))
	}
	if testVerbose {
		testArgs = append(testArgs, "-v")
	}
	if testTimeout > 0 {
		testArgs = append(testArgs, "-timeout", testTimeout.String())
	}
	if testRace {
		testArgs = append(testArgs, "-race")
	}
	testArgs = append(testArgs, testPackage)
	testArgs = append(testArgs, args...)

	pr, pw := io.Pipe()
	analyzer := test.New(pr, util.Output, testVerbose)
	defer func() {
		pw.Close()      // first close pipe writer
		analyzer.Wait() // then quit the read sync loop
	}()

	waitFn, err := util.ExecImmediate("go", testArgs, nil, util.ExecIO{
		Stdout: pw,
		Stderr: pw,
	})
	if err != nil {
		return err
	}
	return waitFn()
}

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
	testCount          int
	testVerbose        bool
	testRace           bool
	testFullPath       bool
	testFailFast       bool
	testBench          bool
	testBenchMem       bool
	testBenchWithTests bool
	testPackage        string
	testRunExpr        string
	testSkipExpr       string
	testBlockProfile   string
	testCoverProfile   string
	testCpuProfile     string
	testMemProfile     string
	testMutextProfile  string
	testCpus           []string
	testTimeout        time.Duration
	testBenchTime      time.Duration
)

var testCommand = &cobra.Command{
	Use:   "test [OPTIONS] [TESTNAME] [-- [ARGS]...]",
	Short: "Test the local package",
	Long:  `Test the local package.`,
	RunE:  runTest,
}

func init() {
	testCommand.Flags().StringVarP(&testRunExpr, "run", "r", "", "Run only those tests matching the regular expression")
	testCommand.Flags().StringVarP(&testSkipExpr, "skip", "s", "", "Skip tests matching the regular expression")
	testCommand.Flags().StringVarP(&testPackage, "package", "p", "./...", "The package to test, default to all packages")
	testCommand.Flags().IntVarP(&testCount, "count", "c", 1, "The number of times to run each test")
	testCommand.Flags().DurationVarP(&testTimeout, "timeout", "t", 0, "The time limit for each test")
	testCommand.Flags().BoolVar(&testRace, "race", false, "Enable race detector")
	testCommand.Flags().BoolVarP(&testVerbose, "verbose", "v", false, "Show output from tests")
	testCommand.Flags().BoolVar(&testFullPath, "full-path", false, "Show full file names in error messages")
	testCommand.Flags().BoolVar(&testFailFast, "fail-fast", false, "Do not start new tests after the first test failure")
	testCommand.Flags().StringSliceVar(&testCpus, "cpu", nil, "Comma-separated list of cpu counts to run each test with")

	testCommand.Flags().BoolVarP(&testBench, "bench", "b", false, "Run only benchmarks matching regexp via --run")
	testCommand.Flags().BoolVar(&testBenchWithTests, "bench-test", false, "Run benchmarks with tests too")
	testCommand.Flags().BoolVar(&testBenchMem, "bench-mem", false, "Print memory allocations for benchmarks")
	testCommand.Flags().DurationVar(&testBenchTime, "bench-time", 0, "Run each benchmark for duration d or N times if `d` is of the form Nx")

	testCommand.Flags().StringVar(&testBlockProfile, "block-profile", "", "Write block profile to file")
	testCommand.Flags().StringVar(&testCoverProfile, "cover-profile", "", "Write coverage profile to file")
	testCommand.Flags().StringVar(&testCpuProfile, "cpu-profile", "", "Write cpu profile to file")
	testCommand.Flags().StringVar(&testMemProfile, "mem-profile", "", "Write memory profile to file")
	testCommand.Flags().StringVar(&testMutextProfile, "mutex-profile", "", "Write mutex profile to file")
}

func runTest(cmd *cobra.Command, args []string) error {
	moduleName, err := util.CurrentModuleName()
	if err != nil {
		return err
	}
	return execGoTest(moduleName, args)
}

func execGoTest(moduleName string, args []string) (err error) {
	var testAllPackages bool
	if strings.HasSuffix(testPackage, allPackagesSuffix) {
		testPackage = strings.TrimSuffix(testPackage, allPackagesSuffix)
		testAllPackages = true
	}

	testPackage, err = parseToGoPackage(moduleName, testPackage)
	if err != nil {
		return
	}
	if testAllPackages {
		testPackage = testPackage + allPackagesSuffix
	}

	var mode test.TestMode
	switch {
	case testBench:
		mode = test.Benchmark
		if testBenchWithTests {
			mode |= test.UnitTest
		}
	default:
		mode = test.UnitTest
	}

	testArgs := buildTestArgs(args)

	errCh := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	pr, pw := io.Pipe()
	analyzer := test.New(pr, moduleName, mode, testCpus, testVerbose)

	go func() {
		defer pw.Close()
		errCh <- util.Exec(ctx, "go", testArgs, nil, util.ExecIO{Stdout: pw, Stderr: pw})
	}()

	analyzer.Wait()
	pr.Close()
	cancel()
	return <-errCh
}

func buildTestArgs(args []string) []string {
	testArgs := []string{"test"}
	if testBench {
		benchExpr := testRunExpr
		if benchExpr == "" {
			benchExpr = "^Benchmark"
		}
		if testBenchWithTests {
			testRunExpr = "^Test"
		} else {
			testRunExpr = "^$" // none
		}
		testArgs = append(testArgs, "-bench", benchExpr)
		if testBenchMem {
			testArgs = append(testArgs, "-benchmem")
		}
		if testBenchTime > 0 {
			testArgs = append(testArgs, "-benchtime", testBenchTime.String())
		}
	}
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
	if testBlockProfile != "" {
		testArgs = append(testArgs, "-blockprofile", testBlockProfile)
	}
	if testCoverProfile != "" {
		testArgs = append(testArgs, "-coverprofile", testCoverProfile)
	}
	if testCpuProfile != "" {
		testArgs = append(testArgs, "-cpuprofile", testCpuProfile)
	}
	if testMemProfile != "" {
		testArgs = append(testArgs, "-memprofile", testMemProfile)
	}
	if testMutextProfile != "" {
		testArgs = append(testArgs, "-mutexprofile", testMutextProfile)
	}
	if len(testCpus) > 0 {
		testArgs = append(testArgs, "-cpu", strings.Join(testCpus, ","))
	}

	testArgs = append(testArgs, "-json")
	testArgs = append(testArgs, testPackage)
	testArgs = append(testArgs, args...)

	return testArgs
}

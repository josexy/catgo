package test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/josexy/catgo/internal/util"
)

const (
	runTestPrefix  = "=== RUN"
	passTestPrefix = "--- PASS:"
	failTestPrefix = "--- FAIL:"
	skipTestPrefix = "--- SKIP:"

	bigPass        = "PASS"
	bigFailNewLine = "FAIL\n"
	noTestPrefix   = "?   \t"
	okPrefix       = "ok  \t"
)

var ignoreSymlinkPrefix = []byte("warning: ignoring symlink")

type TestStatus int

const (
	StatusUnknown TestStatus = iota
	StatusPass
	StatusSkip
	StatusFail
	StatusBuildFail
)

func (status TestStatus) String() string {
	switch status {
	case StatusPass:
		return "PASS"
	case StatusSkip:
		return "SKIPPED"
	case StatusBuildFail, StatusFail:
		return "FAILED"
	}
	return "UNKNOWN"
}

type TestMode uint

const (
	UnitTest TestMode = 1 << iota
	Benchmark
)

type TestEvent struct {
	// encodes as an RFC3339-format string
	Time time.Time
	// start  - the test binary is about to be executed
	// run    - the test has started running
	// pause  - the test has been paused
	// cont   - the test has continued running
	// pass   - the test passed
	// bench  - the benchmark printed log output but did not fail
	// fail   - the test or benchmark failed
	// output - the test printed output
	// skip   - the test was skipped or the package contained no tests
	Action      string
	Package     string
	Test        string
	Elapsed     float64 // seconds
	Output      string
	FailedBuild string
}

type PackageTestEvent struct {
	PackageName string
	Status      TestStatus
	Elapsed     time.Duration
	UnitTests   map[string]*UnitTestEvent
	Benches     map[string]*BenchmarkEvent

	PassedTests  int
	FailedTests  int
	SkippedTests int
}

type BenchmarkEvent struct {
	TestName    string
	Iterations  int64
	NsPerOp     time.Duration
	BytesPerOp  int64
	AllocsPerOp int64
}

type UnitTestEvent struct {
	TestName string
	Time     time.Time
	Status   TestStatus
	Elapsed  time.Duration
}

type LinesOutputAnalyzer struct {
	br *bufio.Reader
	wr io.Writer

	verbose       bool
	moduleName    string
	lastBenchTest string // for bench
	benchBuf      bytes.Buffer
	cpus          []string
	mode          TestMode
	wg            sync.WaitGroup

	events map[string]*PackageTestEvent
}

func New(reader io.Reader, moduleName string, mode TestMode, cpus []string, verbose bool) *LinesOutputAnalyzer {
	if len(cpus) == 0 {
		maxprocs := runtime.GOMAXPROCS(0)
		cpus = []string{strconv.Itoa(maxprocs)}
	}
	analyzer := &LinesOutputAnalyzer{
		moduleName: moduleName,
		verbose:    verbose,
		mode:       mode,
		cpus:       cpus,
		wr:         util.Output,
		br:         bufio.NewReader(reader),
		events:     make(map[string]*PackageTestEvent, 128),
	}
	analyzer.wg.Go(analyzer.runner)
	return analyzer
}

func (a *LinesOutputAnalyzer) Wait() { a.wg.Wait() }

func (a *LinesOutputAnalyzer) runner() {
	var err error
	for {
		var line []byte
		line, _, err = a.br.ReadLine()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			break
		}
		if bytes.HasPrefix(line, ignoreSymlinkPrefix) {
			util.Printer.PrintWarning(string(line))
			continue
		}
		var event TestEvent
		if err = json.Unmarshal(line, &event); err != nil {
			util.Printer.PrintError(fmt.Sprintf("failed to parse line: `%s`, %v", line, err))
			break
		}
		if err = a.analyzeEvent(&event); err != nil {
			util.Printer.PrintError(fmt.Sprintf("failed to analyze event: %v", err))
			break
		}
	}
	if err != nil {
		return
	}
	a.printTestSummary()
}

func (a *LinesOutputAnalyzer) printTestSummary() {
	tw := tabwriter.NewWriter(a.wr, 0, 0, 3, ' ', 0)
	fmt.Fprintln(tw)
	fmt.Fprintln(a.wr, "Test summary:")
	fmt.Fprintln(tw, "PACKAGE\tSTATUS\tPASSED\tFAILED\tSKIPPED\tELAPSED")
	defer tw.Flush()
	var totalPassed, totalFailed, totalSkipped int
	var totalElapsed time.Duration
	for packageName, pv := range a.events {
		totalPassed += pv.PassedTests
		totalFailed += pv.FailedTests
		totalSkipped += pv.SkippedTests
		totalElapsed += pv.Elapsed
		fmt.Fprintf(tw, "%s\t%s\t%d\t%d\t%d\t%s\n",
			formatPackage(packageName, a.moduleName),
			pv.Status.String(),
			pv.PassedTests,
			pv.FailedTests,
			pv.SkippedTests,
			pv.Elapsed.String(),
		)
	}
	fmt.Fprintln(tw)
	fmt.Fprintf(a.wr, "test result: %d passed, %d failed, %d skipped, finished in %s\n\n",
		totalPassed,
		totalFailed,
		totalSkipped,
		totalElapsed.String(),
	)
}

func (a *LinesOutputAnalyzer) analyzeEvent(event *TestEvent) (err error) {
	switch event.Action {
	case "start":
		pv := &PackageTestEvent{
			PackageName: event.Package,
			Status:      StatusUnknown,
			UnitTests:   make(map[string]*UnitTestEvent, 16),
			Benches:     make(map[string]*BenchmarkEvent, 16),
		}
		a.events[event.Package] = pv
		a.printPackageEvent(pv)
	case "run":
		pv, err := a.getPackageTestEvent(event)
		if err != nil {
			return err
		}
		if a.mode&UnitTest == UnitTest {
			pv.UnitTests[event.Test] = &UnitTestEvent{
				TestName: event.Test,
				Time:     event.Time,
				Status:   StatusUnknown,
			}
		}
		a.lastBenchTest = event.Test
		a.printRunningTestEvent(event.Package, event.Test)
	case "build-output", "output":
		if (a.verbose || a.mode&Benchmark == Benchmark) && len(event.Output) > 0 {
			if filterNoneOutputTestLog(event.Output) {
				break
			}
			cont := true
			if a.mode&Benchmark == Benchmark {
				if cont, err = a.tryAnalyzeBenchResult(a.lastBenchTest, event); err != nil {
					return
				}
			}
			if a.verbose && cont {
				fmt.Fprint(a.wr, event.Output)
			}
		}
	case "pass":
		return a.updateEvent(event, StatusPass)
	case "fail":
		return a.updateEvent(event, StatusFail)
	case "skip":
		return a.updateEvent(event, StatusSkip)
	}
	return nil
}

func (a *LinesOutputAnalyzer) tryAnalyzeBenchResult(testName string, event *TestEvent) (bool, error) {
	for _, cpu := range a.cpus {
		// Output looks like:
		// BenchmarkStringConcat-2           174219              6846 ns/op           21080 B/op         99 allocs/op
		// BenchmarkStringConcat-4           133303              7665 ns/op           21080 B/op         99 allocs/op
		//
		// But sometimes the Output may look like, so we have to handle this case with buffer:
		// "BenchmarkStringConcat-2     \t"
		// "  1290205\t       933.7 ns/op\n"
		// "BenchmarkStringConcat-4     \t"
		// "  1290205\t       933.7 ns/op\n"
		if a.benchBuf.Len() > 0 || strings.HasPrefix(event.Output, testName+"-"+cpu+" ") {
			// if the output doesn't end with \n then add it to the buffer
			if event.Output[len(event.Output)-1] != '\n' {
				a.benchBuf.WriteString(event.Output)
				return false, nil
			}

			benchOutput := event.Output
			// if the buffer is not empty mean that the output is split into multiple lines
			if a.benchBuf.Len() > 0 {
				a.benchBuf.WriteString(event.Output)
				benchOutput = a.benchBuf.String()
				a.benchBuf.Reset()
			}

			pv, err := a.getPackageTestEvent(event)
			if err != nil {
				return false, err
			}
			var bench BenchmarkEvent
			fields := strings.Fields(benchOutput)
			if len(fields) > 0 {
				bench.TestName = fields[0] // real test name
			}
			if len(fields) > 1 {
				bench.Iterations, _ = strconv.ParseInt(fields[1], 10, 64)
			}
			if len(fields) > 3 {
				bench.NsPerOp, _ = time.ParseDuration(strings.TrimSuffix(fields[2], " ns/op") + "ns")
			}
			if len(fields) > 5 {
				bench.BytesPerOp, _ = strconv.ParseInt(strings.TrimSuffix(fields[4], " B/op"), 10, 64)
			}
			if len(fields) > 7 {
				bench.AllocsPerOp, _ = strconv.ParseInt(strings.TrimSuffix(fields[6], " allocs/op"), 10, 64)
			}
			if len(bench.TestName) > 0 {
				pv.PassedTests++ // mark as passed...
				pv.Benches[bench.TestName] = &bench
				a.printBenchmarkEventResult(event.Package, &bench)
			}
			return false, nil
		}
	}
	return true, nil
}

func (a *LinesOutputAnalyzer) getPackageTestEvent(event *TestEvent) (*PackageTestEvent, error) {
	pv, ok := a.events[event.Package]
	if !ok {
		return nil, fmt.Errorf("invalid [%s] event for package: %s", event.Action, event.Package)
	}
	return pv, nil
}

func (a *LinesOutputAnalyzer) updateEvent(event *TestEvent, status TestStatus) error {
	pv, err := a.getPackageTestEvent(event)
	if err != nil {
		return err
	}
	if event.Test != "" {
		if uv, ok := pv.UnitTests[event.Test]; ok {
			switch status {
			case StatusPass:
				pv.PassedTests++
			case StatusFail:
				pv.FailedTests++
			case StatusSkip:
				pv.SkippedTests++
			}
			uv.Status = status
			uv.Elapsed = time.Duration(event.Elapsed * float64(time.Second))
			a.printUnitTestEventResult(event.Package, uv)
		}
	} else {
		// treat skip as pass for package level
		if event.Action == "skip" {
			status = StatusPass
		}
		if event.Action == "fail" && len(event.FailedBuild) > 0 {
			status = StatusBuildFail
		}
		pv.Status = status
		pv.Elapsed = time.Duration(event.Elapsed * float64(time.Second))
		a.printPackageEventResult(pv)
	}
	return nil
}

func (a *LinesOutputAnalyzer) printPackageEvent(pv *PackageTestEvent) {
	util.Printer.PrintTesting(fmt.Sprintf("package %s", pv.PackageName))
}

func (a *LinesOutputAnalyzer) printPackageEventResult(pv *PackageTestEvent) {
	util.Printer.BoldGreen.Print("   Finished")
	fmt.Fprintf(a.wr, " test package(%s) result: %s, %d passed, %d failed, %d skipped, finished in %s\n",
		pv.PackageName,
		prettyStatus(pv.Status),
		pv.PassedTests,
		pv.FailedTests,
		pv.SkippedTests,
		pv.Elapsed.String(),
	)
}

func (a *LinesOutputAnalyzer) printUnitTestEventResult(name string, uv *UnitTestEvent) {
	switch uv.Status {
	case StatusPass:
		util.Printer.BoldGreen.Print("     Passed")
	case StatusFail:
		util.Printer.Red.Print("     Failed")
	case StatusSkip:
		util.Printer.Cyan.Print("     Skipped")
	}
	fmt.Fprintf(a.wr, " %s.%s in %s\n", formatPackage(name, a.moduleName), uv.TestName, uv.Elapsed.String())
}

func (a *LinesOutputAnalyzer) printBenchmarkEventResult(name string, be *BenchmarkEvent) {
	util.Printer.BoldGreen.Print("     Done")
	fmt.Fprintf(a.wr, " %s.%s in %d iterations, %s/op, %d B/op, %d allocs/op\n",
		formatPackage(name, a.moduleName), be.TestName, be.Iterations, be.NsPerOp.String(), be.BytesPerOp, be.AllocsPerOp)
}

func (a *LinesOutputAnalyzer) printRunningTestEvent(name, testName string) {
	util.Printer.BoldGreen.Print("     Running")
	fmt.Fprintf(a.wr, " %s.%s\n", formatPackage(name, a.moduleName), testName)
}

func formatPackage(s1, s2 string) string {
	s2 = strings.TrimPrefix(strings.TrimPrefix(s1, s2), "/")
	if s2 == "" {
		return "."
	}
	return s2
}

func prettyStatus(status TestStatus) string {
	switch status {
	case StatusPass:
		return util.Printer.BoldGreen.Sprint(status.String())
	case StatusFail, StatusBuildFail:
		return util.Printer.Red.Sprint(status.String())
	case StatusSkip:
		return util.Printer.Cyan.Sprint(status.String())
	default:
		return util.Printer.Yellow.Sprint(status.String())
	}
}

func filterNoneOutputTestLog(s string) bool {
	return strings.HasPrefix(s, runTestPrefix) ||
		strings.HasPrefix(s, passTestPrefix) ||
		strings.HasPrefix(s, skipTestPrefix) ||
		strings.HasPrefix(s, failTestPrefix) ||
		strings.HasPrefix(s, bigPass) ||
		strings.HasPrefix(s, bigFailNewLine) ||
		strings.HasPrefix(s, noTestPrefix) ||
		strings.HasPrefix(s, okPrefix)
}

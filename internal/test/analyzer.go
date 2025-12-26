package test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/josexy/catgo/internal/util"
)

const (
	runTestPrefix  = "=== RUN"
	passTestPrefix = "--- PASS:"
	failTestPrefix = "--- FAIL:"
	skipTestPrefix = "--- SKIP:"
)

type TestStatus int

const (
	StatusUnknown TestStatus = iota
	StatusPass
	StatusSkip
	StatusFail
	StatusBuildFail
)

type TestResult struct {
	Status        TestStatus
	PassedTests   int
	FailedTests   int
	SkippedTests  int
	HasBuildError bool
	TestDuration  time.Duration
}

type UnitTest struct {
	Name   string
	Status TestStatus
}

type LinesOutputAnalyzer struct {
	br          *bufio.Reader
	wr          io.Writer
	currentLine bytes.Buffer

	status        TestStatus
	passedTests   int
	failedTests   int
	skippedTests  int
	hasBuildError bool
	testDuration  time.Duration

	verbose bool
	wg      sync.WaitGroup
}

func New(reader io.Reader, writer io.Writer, verbose bool) *LinesOutputAnalyzer {
	analyzer := &LinesOutputAnalyzer{
		wr:      writer,
		verbose: verbose,
		br:      bufio.NewReader(reader),
		status:  StatusUnknown,
	}
	analyzer.wg.Go(analyzer.runner)
	return analyzer
}

func (a *LinesOutputAnalyzer) Wait() { a.wg.Wait() }

func (a *LinesOutputAnalyzer) runner() {
	defer func() {
		r := a.GetResult()
		var status string
		if r.IsSuccess() {
			status = util.Printer.Green.Sprint("ok")
		} else {
			status = util.Printer.Red.Sprint("fail")
		}
		fmt.Fprintf(util.Output, "\ntest result: %s. %d passed; %d failed, %d ignored; finished in %s\n",
			status, r.PassedTests, r.FailedTests, r.SkippedTests, r.TestDuration.String())
	}()
	for {
		line, _, err := a.br.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		a.analyzeLine(string(line))
	}
}

func (a *LinesOutputAnalyzer) analyzeLine(line string) {
	defer func() {
		switch a.status {
		case StatusPass:
			util.Printer.Green.Fprintf(a.wr, "%s\n", line)
		case StatusFail, StatusBuildFail:
			util.Printer.Red.Fprintf(a.wr, "%s\n", line)
		case StatusSkip:
			util.Printer.Cyan.Fprintf(a.wr, "%s\n", line)
		case StatusUnknown:
			fmt.Fprintf(a.wr, "%s\n", line)
		}
	}()
	trimmed := strings.TrimSpace(line)

	if strings.HasPrefix(trimmed, "ok") {
		parts := strings.Fields(trimmed)
		if len(parts) > 2 {
			d, _ := time.ParseDuration(parts[2])
			a.testDuration += d
		}
		a.status = StatusPass
		return
	}
	if strings.HasPrefix(trimmed, "PASS") {
		a.status = StatusPass
		return
	}
	if strings.HasPrefix(trimmed, "FAIL") {
		parts := strings.Fields(trimmed)
		if len(parts) > 2 {
			d, _ := time.ParseDuration(parts[2])
			a.testDuration += d
		}
		a.status = StatusFail
		return
	}

	if strings.Contains(line, "can't load package") ||
		strings.Contains(line, "cannot find package") ||
		strings.Contains(line, "undefined:") {
		a.hasBuildError = true
		a.status = StatusBuildFail
		return
	}

	// Check for unit test results
	ut := parseUnitTest(trimmed)
	a.status = ut.Status

	switch ut.Status {
	case StatusPass:
		a.passedTests++
	case StatusFail:
		a.failedTests++
	case StatusSkip:
		a.skippedTests++
	}

	// Check for panic or other critical errors
	if strings.Contains(line, "panic:") || strings.Contains(line, "fatal error:") {
		a.status = StatusFail
	}
}

func parseUnitTest(line string) UnitTest {
	var ut UnitTest
	var parts []string
	if strings.HasPrefix(line, passTestPrefix) || strings.HasPrefix(line, "ok") {
		ut.Status = StatusPass
		parts = strings.Fields(strings.TrimPrefix(line, passTestPrefix))
	} else if strings.HasPrefix(line, failTestPrefix) {
		ut.Status = StatusFail
		parts = strings.Fields(strings.TrimPrefix(line, failTestPrefix))
	} else if strings.HasPrefix(line, skipTestPrefix) {
		ut.Status = StatusSkip
		parts = strings.Fields(strings.TrimPrefix(line, skipTestPrefix))
	}
	if len(parts) > 0 {
		ut.Name = parts[0]
	}
	return ut
}

// GetResult returns the analysis result of all captured output
func (a *LinesOutputAnalyzer) GetResult() *TestResult {
	return &TestResult{
		Status:        a.status,
		PassedTests:   a.passedTests,
		FailedTests:   a.failedTests,
		SkippedTests:  a.skippedTests,
		HasBuildError: a.hasBuildError,
		TestDuration:  a.testDuration,
	}
}

func (r *TestResult) IsSuccess() bool {
	return (r.Status == StatusPass || r.Status == StatusSkip) && r.FailedTests == 0 && !r.HasBuildError
}

func (r *TestResult) IsFailed() bool {
	return r.Status == StatusFail || r.Status == StatusBuildFail || r.FailedTests > 0
}

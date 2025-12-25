package util

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func PathExist(path string) bool {
	info, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist) && info != nil
}

func CurrentDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not get current directory: %w", err)
	}
	return dir, nil
}

func Mkdir(dir string) error {
	if !PathExist(dir) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("could not create directory: %w", err)
		}
	}
	return nil
}

func WriteFile(name string, data []byte) error {
	if err := os.WriteFile(name, data, 0644); err != nil {
		return fmt.Errorf("could not write file %s: %w", name, err)
	}
	return nil
}

func CurrentGoModFile() (string, error) {
	output, err := ExecResult("go", []string{"env", "GOMOD"}, nil)
	if err != nil {
		return "", fmt.Errorf("could not find go.mod file: %w", err)
	}
	goModPath := strings.TrimSpace(string(output))
	if goModPath == "/dev/null" || goModPath == "NUL" || goModPath == "" {
		return "", fmt.Errorf("could not find go.mod file")
	}
	return goModPath, nil
}

func CurrentGoModDir() (string, error) {
	goModPath, err := CurrentGoModFile()
	if err != nil {
		return "", err
	}
	return filepath.Dir(goModPath), nil
}

func CurrentModuleTargetName() (string, error) {
	moduleName, err := CurrentModuleName()
	if err != nil {
		return "", err
	}
	parts := strings.Split(moduleName, "/")
	target := parts[len(parts)-1]
	if runtime.GOOS == "windows" {
		target += ".exe"
	}
	return target, nil
}

func CurrentModuleName() (string, error) {
	goModPath := "go.mod"
	if !PathExist(goModPath) {
		var err error
		if goModPath, err = CurrentGoModFile(); err != nil {
			return "", err
		}
	}
	fp, err := os.Open(goModPath)
	if err != nil {
		return "", fmt.Errorf("could not open go.mod: %w", err)
	}
	defer fp.Close()
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module")), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("could not read go.mod: %w", err)
	}
	return "", fmt.Errorf("module declaration not found in go.mod")
}

func CheckGoInstalled() {
	if value, err := exec.LookPath("go"); err != nil || value == "" {
		Printer.PrintError("go is not installed, please install go first via https://go.dev/doc/install")
		os.Exit(1)
	}
}

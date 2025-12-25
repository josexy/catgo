package util

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := d.Seconds() - float64(minutes*60)
	return fmt.Sprintf("%dm %.2fs", minutes, seconds)
}

func FormatCommandArgs(command string, args []string) string {
	if len(args) == 0 {
		return command
	}
	return command + " " + strings.Join(args, " ")
}

func Exec(command string, args []string, env []string) error {
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = cmd.Environ()
	if len(env) > 0 {
		cmd.Env = append(cmd.Env, env...)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not exec command: `%s`: %w", FormatCommandArgs(command, args), err)
	}
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("process didn't exit successfully: `%s` (exit code: %d)",
				FormatCommandArgs(command, args), exitErr.ExitCode())
		}
		return fmt.Errorf("could not wait for process finish: `%s`: %w", FormatCommandArgs(command, args), err)
	}
	return nil
}

func ExecProcess(command string, args []string, env []string) error {
	if runtime.GOOS == "windows" {
		return Exec(command, args, env)
	}
	cmd := exec.Command(command, args...)
	if cmd.Path == "" && cmd.Err != nil {
		return fmt.Errorf("could not lookup path: `%s`: %w", command, cmd.Err)
	}
	cmd.Env = cmd.Environ()
	if len(env) > 0 {
		cmd.Env = append(cmd.Env, env...)
	}
	if err := syscall.Exec(cmd.Path, cmd.Args, cmd.Env); err != nil {
		return fmt.Errorf("could not exec command: `%s`: %w", FormatCommandArgs(command, args), err)
	}
	return nil
}

func ExecResult(command string, args []string, env []string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	cmd.Env = env
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("could not exec command: `%s`: %w", FormatCommandArgs(command, args), err)
	}
	return output, nil
}

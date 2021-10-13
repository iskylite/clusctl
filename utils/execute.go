package utils

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"
)

func RunShellCmd(cmdStr string) (string, bool) {
	return RunShellCmdWithContext(context.TODO(), cmdStr)
}

func RunShellCmdWithContext(ctx context.Context, cmdStr string) (string, bool) {
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", cmdStr)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return strings.TrimSpace(stderr.String()), cmd.ProcessState.Success()
	}
	return strings.TrimSpace(out.String()), cmd.ProcessState.Success()
}

func ExecuteShellCmd(cmdStr string) (string, bool) {
	return ExecuteShellCmdWithContext(context.TODO(), cmdStr)
}

func ExecuteShellCmdWithContext(ctx context.Context, cmdStr string) (string, bool) {
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", cmdStr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if len(out) == 0 {
			return strings.TrimSpace(err.Error()), false
		}
		return strings.TrimSpace(string(out)) + " " + strings.TrimSpace(err.Error()), false
	}
	return strings.TrimSpace(string(out)), true
}

func ExecuteShellCmdWithTimeout(cmdStr string, timeout int) (string, bool) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Duration(timeout)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", cmdStr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if len(out) == 0 {
			return strings.TrimSpace(err.Error()), false
		}
		return strings.TrimSpace(string(out)), false
	}
	return strings.TrimSpace(string(out)), true
}

func ExecuteShellCmdDaemon(cmdStr string) (*exec.Cmd, error) {
	cmd := exec.Command("/bin/bash", "-c", cmdStr)
	return cmd, cmd.Start()
}

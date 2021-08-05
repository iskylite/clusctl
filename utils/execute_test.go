package utils

import (
	"context"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	out, ok := RunShellCmd("top")
	if ok {
		t.Log(out)
	} else {
		t.Error(out)
	}
}

func TestRunWithContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	out, ok := RunShellCmdWithContext(ctx, "sleep 3 && date")

	if ok {
		t.Log(out)
	} else {
		t.Error(out)
	}
}

func TestExecuteShellCmdWithTimeout(t *testing.T) {
	out, ok := ExecuteShellCmdWithTimeout("sleep 2", 1)
	if ok {
		t.Log(out)
	} else {
		t.Error(out)
	}
}

func TestExecuteShellCmd(t *testing.T) {
	out, ok := ExecuteShellCmd("gg")
	if ok {
		t.Log(out)
	} else {
		t.Error(out)
	}
}

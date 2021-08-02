package utils

import (
	"context"
	"os/exec"
)

func ExecuteShellCmd(cmd string) ([]byte, error) {
	command := exec.Command("/bin/bash", "-c", cmd)
	out, err := command.CombinedOutput()
	return out, err
}

func ExecuteShellCmdWithContext(ctx context.Context, cmd string) ([]byte, error) {
	command := exec.CommandContext(ctx, "/bin/bash", "-c", cmd)
	out, err := command.CombinedOutput()
	return out, err
}

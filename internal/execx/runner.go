package execx

import (
	"context"
	"os"
	"os/exec"
)

type Runner interface {
	Run(ctx context.Context, name string, args ...string) error
	RunElevated(ctx context.Context, name string, args ...string) error
}

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r ExecRunner) RunElevated(ctx context.Context, name string, args ...string) error {
	if os.Geteuid() == 0 {
		return r.Run(ctx, name, args...)
	}

	elevatedArgs := make([]string, 0, len(args)+2)
	elevatedArgs = append(elevatedArgs, "--", name)
	elevatedArgs = append(elevatedArgs, args...)
	return r.Run(ctx, "sudo", elevatedArgs...)
}

package hostpath

import (
	"context"
	"fmt"
	"os/exec"
)

// WSLPath passes the Windows path as one argument, never through a shell.
func WSLPath(ctx context.Context, path string) (string, error) {
	output, err := exec.CommandContext(ctx, "wslpath", "-u", path).Output()
	if err != nil {
		return "", fmt.Errorf("convert Windows path: %w", err)
	}
	return string(output), nil
}

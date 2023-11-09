package shared

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

func StopProcess(ctx context.Context, pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("finding process %d: %w", pid, err)
	}

	if err := proc.Kill(); err != nil {
		slog.Debug("process kill failed, ignoring until we have better process detection")
		return nil
	}

	return nil
}

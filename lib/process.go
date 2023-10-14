package lib

import (
	"fmt"
	"os/exec"
	"runtime"
)

func KillProcess(pid int) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprint(pid))
	case "linux":
		cmd = exec.Command("kill", "-9", fmt.Sprint(pid))
	default:
		return fmt.Errorf("Unsupported operating system: %s", runtime.GOOS)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to kill process with PID %d: %w", pid, err)
	}

	return nil
}

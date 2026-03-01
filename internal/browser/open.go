package browser

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OpenURL opens a URL with the system-default handler.
func OpenURL(targetURL string) error {
	cmd, err := openCommand(runtime.GOOS, targetURL)
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start browser: %w", err)
	}

	return nil
}

func openCommand(goos, targetURL string) (*exec.Cmd, error) {
	switch goos {
	case "darwin":
		return exec.Command("open", targetURL), nil
	case "linux":
		return exec.Command("xdg-open", targetURL), nil
	case "windows":
		// Avoid `cmd /c start` because it is shell-interpreted.
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL), nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", goos)
	}
}

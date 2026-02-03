package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/dedene/raindrop-cli/internal/errfmt"
)

type OpenCmd struct {
	ID int `arg:"" help:"Raindrop ID"`
}

func (c *OpenCmd) Run(_ *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	raindrop, err := client.GetRaindrop(ctx, c.ID)
	if err != nil {
		return errfmt.Format(err)
	}

	return openBrowser(raindrop.Link)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}

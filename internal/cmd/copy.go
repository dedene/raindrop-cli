package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/dedene/raindrop-cli/internal/errfmt"
)

type CopyCmd struct {
	ID int `arg:"" help:"Raindrop ID"`
}

func (c *CopyCmd) Run(_ *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	raindrop, err := client.GetRaindrop(ctx, c.ID)
	if err != nil {
		return errfmt.Format(err)
	}

	if err := copyToClipboard(raindrop.Link); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Copied: %s\n", raindrop.Link)

	return nil
}

func copyToClipboard(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case "windows":
		cmd = exec.Command("clip")
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("get stdin: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	if _, err := stdin.Write([]byte(text)); err != nil {
		return fmt.Errorf("write to clipboard: %w", err)
	}

	if err := stdin.Close(); err != nil {
		return fmt.Errorf("close stdin: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("wait for command: %w", err)
	}

	return nil
}

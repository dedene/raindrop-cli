package cmd

import (
	"fmt"
	"os"

	"github.com/dedene/raindrop-cli/internal/errfmt"
)

type DeleteCmd struct {
	ID        int  `arg:"" help:"Raindrop ID"`
	Permanent bool `help:"Permanently delete (skip trash)" short:"p"`
}

func (c *DeleteCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	// Get raindrop first to show what we're deleting
	raindrop, err := client.GetRaindrop(ctx, c.ID)
	if err != nil {
		return errfmt.Format(err)
	}

	action := "move to trash"
	if c.Permanent {
		action = "permanently delete"
	}

	msg := fmt.Sprintf("%s '%s' (ID: %d)?", action, truncate(raindrop.Title, 40), c.ID)
	if !confirmAction(msg, flags) {
		fmt.Fprintln(os.Stdout, "Cancelled.")

		return nil
	}

	if err := client.DeleteRaindrop(ctx, c.ID, c.Permanent); err != nil {
		return errfmt.Format(err)
	}

	if c.Permanent {
		fmt.Fprintln(os.Stdout, "Permanently deleted.")
	} else {
		fmt.Fprintln(os.Stdout, "Moved to trash.")
	}

	return nil
}

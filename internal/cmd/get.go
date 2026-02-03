package cmd

import (
	"os"

	"github.com/dedene/raindrop-cli/internal/errfmt"
	"github.com/dedene/raindrop-cli/internal/output"
)

type GetCmd struct {
	ID int `arg:"" help:"Raindrop ID"`
}

func (c *GetCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	raindrop, err := client.GetRaindrop(ctx, c.ID)
	if err != nil {
		return errfmt.Format(err)
	}

	if flags.JSON {
		return output.WriteJSON(os.Stdout, raindrop)
	}

	output.FormatRaindropDetail(os.Stdout, raindrop)

	return nil
}

package cmd

import (
	"github.com/dedene/raindrop-cli/internal/browser"
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
	return browser.OpenURL(url)
}

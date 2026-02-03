package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dedene/raindrop-cli/internal/api"
	"github.com/dedene/raindrop-cli/internal/errfmt"
	"github.com/dedene/raindrop-cli/internal/output"
)

type UpdateCmd struct {
	ID         int      `arg:"" help:"Raindrop ID"`
	Title      string   `help:"New title" short:"t"`
	Collection string   `help:"Move to collection" short:"c"`
	Tags       []string `help:"Replace tags" short:"T"`
	Note       string   `help:"Update note" short:"n"`
	Favorite   bool     `help:"Mark as favorite" short:"f"`
	Unfavorite bool     `help:"Remove favorite"`
}

func (c *UpdateCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	req := &api.UpdateRaindropRequest{}
	hasChanges := false

	if c.Title != "" {
		req.Title = c.Title
		hasChanges = true
	}

	if c.Note != "" {
		req.Note = c.Note
		hasChanges = true
	}

	if len(c.Tags) > 0 {
		req.Tags = c.normalizeTags()
		hasChanges = true
	}

	if c.Favorite {
		v := true
		req.Important = &v
		hasChanges = true
	}

	if c.Unfavorite {
		v := false
		req.Important = &v
		hasChanges = true
	}

	if c.Collection != "" {
		collectionID, resolveErr := client.ResolveCollection(ctx, c.Collection)
		if resolveErr != nil {
			return errfmt.Format(resolveErr)
		}

		req.Collection = &struct {
			ID int `json:"$id"`
		}{ID: collectionID}
		hasChanges = true
	}

	if !hasChanges {
		return fmt.Errorf("no changes specified; use --title, --tags, --note, --collection, --favorite, or --unfavorite")
	}

	raindrop, err := client.UpdateRaindrop(ctx, c.ID, req)
	if err != nil {
		return errfmt.Format(err)
	}

	if flags.JSON {
		return output.WriteJSON(os.Stdout, raindrop)
	}

	fmt.Fprintf(os.Stdout, "Updated: %s (ID: %d)\n", raindrop.Title, raindrop.ID)

	return nil
}

func (c *UpdateCmd) normalizeTags() []string {
	var tags []string

	for _, t := range c.Tags {
		for _, part := range strings.Split(t, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				tags = append(tags, part)
			}
		}
	}

	return tags
}

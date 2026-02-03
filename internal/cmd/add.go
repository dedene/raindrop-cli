package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dedene/raindrop-cli/internal/api"
	"github.com/dedene/raindrop-cli/internal/errfmt"
	"github.com/dedene/raindrop-cli/internal/output"
)

type AddCmd struct {
	URL        string   `arg:"" optional:"" help:"URL to add (use - for stdin bulk)"`
	Collection string   `help:"Collection name or ID" default:"-1" short:"c"`
	Title      string   `help:"Override title" short:"t"`
	Tags       []string `help:"Tags (repeat flag or comma-separated)" short:"T"`
	Note       string   `help:"Note text" short:"n"`
	NoFetch    bool     `help:"Skip fetching URL metadata" name:"no-fetch"`
}

func (c *AddCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	collectionID, err := client.ResolveCollection(ctx, c.Collection)
	if err != nil {
		return errfmt.Format(err)
	}

	// Handle stdin bulk input
	if c.URL == "-" {
		return c.runBulk(client, flags, collectionID)
	}

	if c.URL == "" {
		return fmt.Errorf("URL is required (use - for stdin)")
	}

	// Parse URL for metadata if not skipped and no title override
	var parsed *api.ParsedURL

	if !c.NoFetch && c.Title == "" {
		parsed, _ = client.ParseURL(ctx, c.URL) // ignore errors, continue without metadata
	}

	req := &api.CreateRaindropRequest{
		Link:        c.URL,
		Tags:        c.normalizeTags(),
		Note:        c.Note,
		PleaseParse: !c.NoFetch && c.Title == "",
	}
	req.Collection.ID = collectionID

	if c.Title != "" {
		req.Title = c.Title
	} else if parsed != nil && parsed.Item.Title != "" {
		req.Title = parsed.Item.Title
	}

	raindrop, err := client.CreateRaindrop(ctx, req)
	if err != nil {
		return errfmt.Format(err)
	}

	if flags.JSON {
		return output.WriteJSON(os.Stdout, raindrop)
	}

	fmt.Fprintf(os.Stdout, "Added: %s (ID: %d)\n", raindrop.Title, raindrop.ID)

	return nil
}

func (c *AddCmd) runBulk(client *api.Client, flags *RootFlags, collectionID int) error {
	urls, err := readURLsFromStdin()
	if err != nil {
		return err
	}

	if len(urls) == 0 {
		return fmt.Errorf("no URLs provided on stdin")
	}

	_, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	// Build requests
	items := make([]api.CreateRaindropRequest, 0, len(urls))

	for _, u := range urls {
		req := api.CreateRaindropRequest{
			Link:        u,
			Tags:        c.normalizeTags(),
			PleaseParse: true,
		}
		req.Collection.ID = collectionID

		items = append(items, req)
	}

	// Batch in groups of 100
	var allCreated []api.Raindrop

	batchSize := 100

	for i := 0; i < len(items); i += batchSize {
		end := min(i+batchSize, len(items))

		batch := items[i:end]

		created, err := client.CreateRaindropsBulk(ctx, batch)
		if err != nil {
			return errfmt.Format(err)
		}

		allCreated = append(allCreated, created...)

		if !flags.JSON {
			fmt.Fprintf(os.Stderr, "Created %d/%d raindrops\n", len(allCreated), len(items))
		}
	}

	if flags.JSON {
		return output.WriteJSON(os.Stdout, allCreated)
	}

	fmt.Fprintf(os.Stdout, "Added %d raindrops\n", len(allCreated))

	return nil
}

func (c *AddCmd) normalizeTags() []string {
	var tags []string

	for _, t := range c.Tags {
		// Handle comma-separated values
		for _, part := range strings.Split(t, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				tags = append(tags, part)
			}
		}
	}

	return tags
}

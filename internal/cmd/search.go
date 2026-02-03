package cmd

import (
	"fmt"
	"os"

	"github.com/dedene/raindrop-cli/internal/api"
	"github.com/dedene/raindrop-cli/internal/errfmt"
	"github.com/dedene/raindrop-cli/internal/output"
)

type SearchCmd struct {
	Query      string `arg:"" optional:"" help:"Search query"`
	Tag        string `help:"Filter by tag" short:"t"`
	Type       string `help:"Filter by type (link|article|image|video|document|audio)" short:"T"`
	After      string `help:"Created after date (YYYY-MM-DD)"`
	Before     string `help:"Created before date (YYYY-MM-DD)"`
	Collection string `help:"Collection to search" default:"0" short:"c"`
	All        bool   `help:"Fetch all results" short:"a"`
}

func (c *SearchCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	collectionID, err := client.ResolveCollection(ctx, c.Collection)
	if err != nil {
		return errfmt.Format(err)
	}

	search := c.buildSearch()
	if search == "" {
		return fmt.Errorf("search query required; use positional arg or --tag, --type, --after, --before")
	}

	opts := api.ListOptions{
		Search:  search,
		Sort:    "score", // relevance for search
		PerPage: 50,
	}

	var allItems []api.Raindrop

	page := 0

	for {
		opts.Page = page

		resp, err := client.ListRaindrops(ctx, collectionID, opts)
		if err != nil {
			return errfmt.Format(err)
		}

		allItems = append(allItems, resp.Items...)

		if !c.All || len(allItems) >= resp.Count {
			break
		}

		page++
	}

	if flags.JSON {
		return output.WriteJSON(os.Stdout, allItems)
	}

	if len(allItems) == 0 {
		fmt.Fprintln(os.Stdout, "No results found.")

		return nil
	}

	tw := output.NewTableWriter(os.Stdout, output.RaindropTableHeaders()...)
	for _, r := range allItems {
		tw.AddRow(output.FormatRaindropRow(&r, flags.HyperlinkMode())...)
	}

	tw.Render()
	fmt.Fprintf(os.Stdout, "\n%d result(s)\n", len(allItems))

	return nil
}

func (c *SearchCmd) buildSearch() string {
	var parts []string

	if c.Query != "" {
		parts = append(parts, c.Query)
	}

	if c.Tag != "" {
		parts = append(parts, "#"+c.Tag)
	}

	if c.Type != "" {
		parts = append(parts, "type:"+c.Type)
	}

	if c.After != "" {
		parts = append(parts, "created:>"+c.After)
	}

	if c.Before != "" {
		parts = append(parts, "created:<"+c.Before)
	}

	return joinSearchParts(parts)
}

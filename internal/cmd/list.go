package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dedene/raindrop-cli/internal/api"
	"github.com/dedene/raindrop-cli/internal/errfmt"
	"github.com/dedene/raindrop-cli/internal/output"
)

type ListCmd struct {
	Collection string `arg:"" optional:"" help:"Collection name/ID (default: all)" default:"0"`
	Favorites  bool   `help:"Only favorites" short:"f"`
	Broken     bool   `help:"Only broken links"`
	Type       string `help:"Filter by type (link|article|image|video|document|audio)" short:"t"`
	Tag        string `help:"Filter by tag"`
	Search     string `help:"Search query" short:"s"`
	Sort       string `help:"Sort order" default:"-created" enum:"created,-created,title,-title,domain,-domain,score"`
	All        bool   `help:"Fetch all pages (default: first 50)" short:"a"`
}

func (c *ListCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	collectionID, err := client.ResolveCollection(ctx, c.Collection)
	if err != nil {
		return errfmt.Format(err)
	}

	// Build search query from filters
	search := c.buildSearch()

	opts := api.ListOptions{
		Search:  search,
		Sort:    c.Sort,
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

		// Stop if not fetching all or no more items
		if !c.All || len(allItems) >= resp.Count {
			break
		}

		page++
	}

	if flags.JSON {
		return output.WriteJSON(os.Stdout, allItems)
	}

	if len(allItems) == 0 {
		fmt.Fprintln(os.Stdout, "No raindrops found.")

		return nil
	}

	tw := output.NewTableWriter(os.Stdout, output.RaindropTableHeaders()...)
	for _, r := range allItems {
		tw.AddRow(output.FormatRaindropRow(&r, flags.HyperlinkMode())...)
	}

	tw.Render()
	fmt.Fprintf(os.Stdout, "\n%d raindrop(s)\n", len(allItems))

	return nil
}

func (c *ListCmd) buildSearch() string {
	var parts []string

	if c.Search != "" {
		parts = append(parts, c.Search)
	}

	if c.Tag != "" {
		parts = append(parts, "#"+c.Tag)
	}

	if c.Type != "" {
		parts = append(parts, "type:"+c.Type)
	}

	if c.Favorites {
		parts = append(parts, "important:true")
	}

	if c.Broken {
		parts = append(parts, "broken:true")
	}

	if len(parts) == 0 {
		return ""
	}

	return joinSearchParts(parts)
}

func joinSearchParts(parts []string) string {
	return strings.Join(parts, " ")
}

package cmd

import (
	"fmt"
	"os"

	"github.com/dedene/raindrop-cli/internal/api"
	"github.com/dedene/raindrop-cli/internal/errfmt"
	"github.com/dedene/raindrop-cli/internal/output"
)

type HighlightsCmd struct {
	List   HighlightsListCmd   `cmd:"" help:"List highlights for a raindrop"`
	Add    HighlightsAddCmd    `cmd:"" help:"Add a highlight"`
	Delete HighlightsDeleteCmd `cmd:"" help:"Delete a highlight"`
}

type HighlightsListCmd struct {
	ID int `arg:"" help:"Raindrop ID"`
}

func (c *HighlightsListCmd) Run(flags *RootFlags) error {
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
		return output.WriteJSON(os.Stdout, raindrop.Highlights)
	}

	if len(raindrop.Highlights) == 0 {
		fmt.Fprintln(os.Stdout, "No highlights found.")

		return nil
	}

	for i, h := range raindrop.Highlights {
		fmt.Fprintf(os.Stdout, "%s %d. %s\n", output.StyleBold("Highlight"), i+1, h.Text)

		if h.Note != "" {
			fmt.Fprintf(os.Stdout, "   %s %s\n", output.StyleFaint("Note:"), h.Note)
		}

		fmt.Fprintf(os.Stdout, "   %s %s  %s %s\n",
			output.StyleFaint("Color:"), h.Color,
			output.StyleFaint("ID:"), h.ID)
		fmt.Fprintln(os.Stdout)
	}

	fmt.Fprintf(os.Stdout, "%d highlight(s)\n", len(raindrop.Highlights))

	return nil
}

type HighlightsAddCmd struct {
	RaindropID int    `arg:"" help:"Raindrop ID"`
	Text       string `arg:"" help:"Highlight text"`
	Note       string `help:"Annotation" short:"n"`
	Color      string `help:"Color" enum:"yellow,blue,red,green,purple" default:"yellow" short:"c"`
}

func (c *HighlightsAddCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	// Get current raindrop to append highlight
	raindrop, err := client.GetRaindrop(ctx, c.RaindropID)
	if err != nil {
		return errfmt.Format(err)
	}

	// Create new highlight
	newHighlight := api.Highlight{
		Text:  c.Text,
		Note:  c.Note,
		Color: c.Color,
	}

	// Append to existing highlights
	highlights := make([]api.Highlight, len(raindrop.Highlights)+1)
	copy(highlights, raindrop.Highlights)
	highlights[len(raindrop.Highlights)] = newHighlight

	// Update raindrop with new highlights via raw request
	// The API expects highlights array in PUT /raindrop/{id}
	type updateReq struct {
		Highlights []api.Highlight `json:"highlights"`
	}

	req := updateReq{Highlights: highlights}

	var resp struct {
		Item api.Raindrop `json:"item"`
	}

	if err := client.Put(ctx, fmt.Sprintf("/raindrop/%d", c.RaindropID), &req, &resp); err != nil {
		return errfmt.Format(err)
	}

	if flags.JSON {
		return output.WriteJSON(os.Stdout, resp.Item.Highlights)
	}

	fmt.Fprintf(os.Stdout, "Added highlight to '%s'\n", resp.Item.Title)

	return nil
}

type HighlightsDeleteCmd struct {
	RaindropID  int    `arg:"" help:"Raindrop ID"`
	HighlightID string `arg:"" help:"Highlight ID"`
}

func (c *HighlightsDeleteCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	// Get current raindrop
	raindrop, err := client.GetRaindrop(ctx, c.RaindropID)
	if err != nil {
		return errfmt.Format(err)
	}

	// Filter out the highlight to delete
	remaining := make([]api.Highlight, 0, len(raindrop.Highlights))
	found := false

	for _, h := range raindrop.Highlights {
		if h.ID == c.HighlightID {
			found = true

			continue
		}

		remaining = append(remaining, h)
	}

	if !found {
		return fmt.Errorf("highlight not found: %s", c.HighlightID)
	}

	// Update raindrop with remaining highlights
	type updateReq struct {
		Highlights []api.Highlight `json:"highlights"`
	}

	req := updateReq{Highlights: remaining}

	var resp struct {
		Item api.Raindrop `json:"item"`
	}

	if err := client.Put(ctx, fmt.Sprintf("/raindrop/%d", c.RaindropID), &req, &resp); err != nil {
		return errfmt.Format(err)
	}

	fmt.Fprintln(os.Stdout, "Deleted highlight.")

	return nil
}

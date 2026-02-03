package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/dedene/raindrop-cli/internal/api"
)

// FormatRaindropRow returns table row columns for a raindrop.
func FormatRaindropRow(r *api.Raindrop, hyperlinkMode HyperlinkMode) []string {
	title := r.Title
	if len(title) > 50 {
		title = title[:47] + "..."
	}

	// URL column - truncate display but link to full URL when supported
	const maxURLWidth = 40
	displayURL := TruncateURL(r.Link, maxURLWidth)
	urlCell := MaybeHyperlink(r.Link, displayURL, hyperlinkMode)

	return []string{
		fmt.Sprintf("%d", r.ID),
		title,
		urlCell,
		r.Type,
		r.Created.Format("2006-01-02 15:04"),
	}
}

// FormatRaindropDetail writes full raindrop details to w.
func FormatRaindropDetail(w io.Writer, r *api.Raindrop) {
	fmt.Fprintf(w, "%s %d\n", StyleBold("ID:"), r.ID)
	fmt.Fprintf(w, "%s %s\n", StyleBold("Title:"), r.Title)
	fmt.Fprintf(w, "%s %s\n", StyleBold("URL:"), r.Link)
	fmt.Fprintf(w, "%s %s\n", StyleBold("Domain:"), r.Domain)
	fmt.Fprintf(w, "%s %s\n", StyleBold("Type:"), r.Type)

	if r.Excerpt != "" {
		fmt.Fprintf(w, "%s %s\n", StyleBold("Excerpt:"), r.Excerpt)
	}

	if len(r.Tags) > 0 {
		fmt.Fprintf(w, "%s %s\n", StyleBold("Tags:"), strings.Join(r.Tags, ", "))
	}

	if r.Important {
		fmt.Fprintf(w, "%s %s\n", StyleBold("Favorite:"), StyleYellow("yes"))
	}

	if r.Note != "" {
		fmt.Fprintf(w, "%s\n%s\n", StyleBold("Note:"), r.Note)
	}

	fmt.Fprintf(w, "%s %s\n", StyleBold("Created:"), r.Created.Format("2006-01-02 15:04"))
	fmt.Fprintf(w, "%s %s\n", StyleBold("Updated:"), r.Updated.Format("2006-01-02 15:04"))

	if len(r.Highlights) > 0 {
		fmt.Fprintf(w, "\n%s\n", StyleBold("Highlights:"))

		for i, h := range r.Highlights {
			fmt.Fprintf(w, "  %d. %s\n", i+1, h.Text)

			if h.Note != "" {
				fmt.Fprintf(w, "     %s %s\n", StyleFaint("Note:"), h.Note)
			}
		}
	}
}

// RaindropTableHeaders returns standard headers for raindrop list.
func RaindropTableHeaders() []string {
	return []string{"ID", "TITLE", "URL", "TYPE", "CREATED"}
}

package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/dedene/raindrop-cli/internal/api"
)

// FormatRaindropRow returns table row columns for a raindrop.
func FormatRaindropRow(r *api.Raindrop, hyperlinkMode HyperlinkMode) []string {
	title := SanitizeInline(r.Title)
	if len(title) > 50 {
		title = title[:47] + "..."
	}

	// URL column - truncate display but link to full URL when supported
	const maxURLWidth = 40
	safeLink := SanitizeInline(r.Link)
	displayURL := TruncateURL(safeLink, maxURLWidth)
	urlCell := MaybeHyperlink(safeLink, displayURL, hyperlinkMode)

	return []string{
		fmt.Sprintf("%d", r.ID),
		title,
		urlCell,
		SanitizeInline(r.Type),
		r.Created.Format("2006-01-02 15:04"),
	}
}

// FormatRaindropDetail writes full raindrop details to w.
func FormatRaindropDetail(w io.Writer, r *api.Raindrop) {
	fmt.Fprintf(w, "%s %d\n", StyleBold("ID:"), r.ID)
	fmt.Fprintf(w, "%s %s\n", StyleBold("Title:"), SanitizeInline(r.Title))
	fmt.Fprintf(w, "%s %s\n", StyleBold("URL:"), SanitizeInline(r.Link))
	fmt.Fprintf(w, "%s %s\n", StyleBold("Domain:"), SanitizeInline(r.Domain))
	fmt.Fprintf(w, "%s %s\n", StyleBold("Type:"), SanitizeInline(r.Type))

	if r.Excerpt != "" {
		fmt.Fprintf(w, "%s %s\n", StyleBold("Excerpt:"), SanitizeInline(r.Excerpt))
	}

	if len(r.Tags) > 0 {
		tags := make([]string, 0, len(r.Tags))
		for _, tag := range r.Tags {
			tags = append(tags, SanitizeInline(tag))
		}
		fmt.Fprintf(w, "%s %s\n", StyleBold("Tags:"), strings.Join(tags, ", "))
	}

	if r.Important {
		fmt.Fprintf(w, "%s %s\n", StyleBold("Favorite:"), StyleYellow("yes"))
	}

	if r.Note != "" {
		fmt.Fprintf(w, "%s\n%s\n", StyleBold("Note:"), SanitizeText(r.Note))
	}

	fmt.Fprintf(w, "%s %s\n", StyleBold("Created:"), r.Created.Format("2006-01-02 15:04"))
	fmt.Fprintf(w, "%s %s\n", StyleBold("Updated:"), r.Updated.Format("2006-01-02 15:04"))

	if len(r.Highlights) > 0 {
		fmt.Fprintf(w, "\n%s\n", StyleBold("Highlights:"))

		for i, h := range r.Highlights {
			fmt.Fprintf(w, "  %d. %s\n", i+1, SanitizeText(h.Text))

			if h.Note != "" {
				fmt.Fprintf(w, "     %s %s\n", StyleFaint("Note:"), SanitizeText(h.Note))
			}
		}
	}
}

// RaindropTableHeaders returns standard headers for raindrop list.
func RaindropTableHeaders() []string {
	return []string{"ID", "TITLE", "URL", "TYPE", "CREATED"}
}

package output

import (
	"fmt"
	"io"
	"sort"

	"github.com/dedene/raindrop-cli/internal/api"
)

// FormatCollectionRow returns table row columns for a collection.
func FormatCollectionRow(c *api.Collection) []string {
	parent := ""
	if c.ParentID() != 0 {
		parent = fmt.Sprintf("%d", c.ParentID())
	}

	return []string{
		fmt.Sprintf("%d", c.ID),
		c.Title,
		fmt.Sprintf("%d", c.Count),
		parent,
	}
}

// CollectionTableHeaders returns standard headers for collection list.
func CollectionTableHeaders() []string {
	return []string{"ID", "NAME", "COUNT", "PARENT"}
}

// FormatCollectionDetail writes full collection details to w.
func FormatCollectionDetail(w io.Writer, c *api.Collection) {
	fmt.Fprintf(w, "%s %d\n", StyleBold("ID:"), c.ID)
	fmt.Fprintf(w, "%s %s\n", StyleBold("Name:"), c.Title)
	fmt.Fprintf(w, "%s %d\n", StyleBold("Count:"), c.Count)

	if c.ParentID() != 0 {
		fmt.Fprintf(w, "%s %d\n", StyleBold("Parent ID:"), c.ParentID())
	}

	if c.Color != "" {
		fmt.Fprintf(w, "%s %s\n", StyleBold("Color:"), c.Color)
	}

	fmt.Fprintf(w, "%s %s\n", StyleBold("Created:"), c.Created.Format("2006-01-02 15:04"))
	fmt.Fprintf(w, "%s %s\n", StyleBold("Updated:"), c.Updated.Format("2006-01-02 15:04"))
}

// CollectionTree renders collections as a tree structure.
type CollectionTree struct {
	w           io.Writer
	collections []api.Collection
	byParent    map[int][]api.Collection
}

// NewCollectionTree creates a tree renderer.
func NewCollectionTree(w io.Writer, collections []api.Collection) *CollectionTree {
	t := &CollectionTree{
		w:           w,
		collections: collections,
		byParent:    make(map[int][]api.Collection),
	}

	// Group by parent
	for _, c := range collections {
		t.byParent[c.ParentID()] = append(t.byParent[c.ParentID()], c)
	}

	// Sort each group alphabetically
	for k := range t.byParent {
		sort.Slice(t.byParent[k], func(i, j int) bool {
			return t.byParent[k][i].Title < t.byParent[k][j].Title
		})
	}

	return t
}

// Render writes the tree to the output.
func (t *CollectionTree) Render() {
	// Print system collections first
	fmt.Fprintf(t.w, "%s\n", StyleFaint("System:"))
	fmt.Fprintf(t.w, "  All (0)\n")
	fmt.Fprintf(t.w, "  Unsorted (-1)\n")
	fmt.Fprintf(t.w, "  Trash (-99)\n")
	fmt.Fprintln(t.w)

	if len(t.collections) == 0 {
		fmt.Fprintf(t.w, "%s\n", StyleFaint("No custom collections"))

		return
	}

	fmt.Fprintf(t.w, "%s\n", StyleFaint("Collections:"))
	t.renderChildren(0, "")
}

func (t *CollectionTree) renderChildren(parentID int, prefix string) {
	children := t.byParent[parentID]

	for i, c := range children {
		isLast := i == len(children)-1
		connector := "├── "

		if isLast {
			connector = "└── "
		}

		fmt.Fprintf(t.w, "%s%s%s (%d)\n", prefix, connector, c.Title, c.Count)

		// Recurse for children
		childPrefix := prefix + "│   "
		if isLast {
			childPrefix = prefix + "    "
		}

		t.renderChildren(c.ID, childPrefix)
	}
}

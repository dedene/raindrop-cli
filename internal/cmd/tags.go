package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dedene/raindrop-cli/internal/errfmt"
	"github.com/dedene/raindrop-cli/internal/output"
)

type TagsCmd struct {
	List   TagsListCmd   `cmd:"" default:"1" help:"List tags"`
	Rename TagsRenameCmd `cmd:"" help:"Rename a tag"`
	Merge  TagsMergeCmd  `cmd:"" help:"Merge tags into one"`
	Delete TagsDeleteCmd `cmd:"" help:"Delete tags"`
}

type TagsListCmd struct {
	Collection string `help:"Collection name/ID (default: all)" default:"0" short:"c"`
}

func (c *TagsListCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	collectionID, err := client.ResolveCollection(ctx, c.Collection)
	if err != nil {
		return errfmt.Format(err)
	}

	tags, err := client.ListTags(ctx, collectionID)
	if err != nil {
		return errfmt.Format(err)
	}

	if flags.JSON {
		return output.WriteJSON(os.Stdout, tags)
	}

	if len(tags) == 0 {
		fmt.Fprintln(os.Stdout, "No tags found.")

		return nil
	}

	tw := output.NewTableWriter(os.Stdout, "TAG", "COUNT")
	for _, t := range tags {
		tw.AddRow(t.Tag, fmt.Sprintf("%d", t.Count))
	}

	tw.Render()
	fmt.Fprintf(os.Stdout, "\n%d tag(s)\n", len(tags))

	return nil
}

type TagsRenameCmd struct {
	Old        string `arg:"" help:"Current tag name"`
	New        string `arg:"" help:"New tag name"`
	Collection string `help:"Collection name/ID (default: all)" default:"0" short:"c"`
}

func (c *TagsRenameCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	collectionID, err := client.ResolveCollection(ctx, c.Collection)
	if err != nil {
		return errfmt.Format(err)
	}

	if err := client.RenameTags(ctx, collectionID, []string{c.Old}, c.New); err != nil {
		return errfmt.Format(err)
	}

	fmt.Fprintf(os.Stdout, "Renamed '%s' to '%s'\n", c.Old, c.New)

	return nil
}

type TagsMergeCmd struct {
	Tags       string `arg:"" help:"Tags to merge (comma-separated)"`
	Into       string `help:"Target tag name" required:""`
	Collection string `help:"Collection name/ID (default: all)" default:"0" short:"c"`
}

func (c *TagsMergeCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	collectionID, err := client.ResolveCollection(ctx, c.Collection)
	if err != nil {
		return errfmt.Format(err)
	}

	// Parse comma-separated tags
	var tags []string

	for _, t := range strings.Split(c.Tags, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}

	if len(tags) == 0 {
		return fmt.Errorf("no tags specified")
	}

	if err := client.RenameTags(ctx, collectionID, tags, c.Into); err != nil {
		return errfmt.Format(err)
	}

	fmt.Fprintf(os.Stdout, "Merged %d tags into '%s'\n", len(tags), c.Into)

	return nil
}

type TagsDeleteCmd struct {
	Tags       string `arg:"" help:"Tags to delete (comma-separated)"`
	Collection string `help:"Collection name/ID (default: all)" default:"0" short:"c"`
}

func (c *TagsDeleteCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	collectionID, err := client.ResolveCollection(ctx, c.Collection)
	if err != nil {
		return errfmt.Format(err)
	}

	// Parse comma-separated tags
	var tags []string

	for _, t := range strings.Split(c.Tags, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}

	if len(tags) == 0 {
		return fmt.Errorf("no tags specified")
	}

	msg := fmt.Sprintf("Delete %d tag(s)?", len(tags))
	if !confirmAction(msg, flags) {
		fmt.Fprintln(os.Stdout, "Cancelled.")

		return nil
	}

	if err := client.DeleteTags(ctx, collectionID, tags); err != nil {
		return errfmt.Format(err)
	}

	fmt.Fprintf(os.Stdout, "Deleted %d tag(s)\n", len(tags))

	return nil
}

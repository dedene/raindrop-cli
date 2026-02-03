package cmd

import (
	"fmt"
	"os"

	"github.com/dedene/raindrop-cli/internal/api"
	"github.com/dedene/raindrop-cli/internal/errfmt"
	"github.com/dedene/raindrop-cli/internal/output"
)

type CollectionsCmd struct {
	List   CollectionsListCmd   `cmd:"" default:"1" help:"List collections"`
	Get    CollectionsGetCmd    `cmd:"" help:"Get collection details"`
	Create CollectionsCreateCmd `cmd:"" help:"Create a collection"`
	Update CollectionsUpdateCmd `cmd:"" help:"Update a collection"`
	Delete CollectionsDeleteCmd `cmd:"" help:"Delete a collection"`
}

type CollectionsListCmd struct {
	Flat bool `help:"Flat list instead of tree" short:"f"`
}

func (c *CollectionsListCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	collections, err := client.ListAllCollections(ctx)
	if err != nil {
		return errfmt.Format(err)
	}

	if flags.JSON {
		return output.WriteJSON(os.Stdout, collections)
	}

	if c.Flat {
		tw := output.NewTableWriter(os.Stdout, output.CollectionTableHeaders()...)
		for _, col := range collections {
			tw.AddRow(output.FormatCollectionRow(&col)...)
		}

		tw.Render()
		fmt.Fprintf(os.Stdout, "\n%d collection(s)\n", len(collections))

		return nil
	}

	// Tree view
	tree := output.NewCollectionTree(os.Stdout, collections)
	tree.Render()

	return nil
}

type CollectionsGetCmd struct {
	Collection string `arg:"" help:"Collection name or ID"`
}

func (c *CollectionsGetCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	collectionID, err := client.ResolveCollection(ctx, c.Collection)
	if err != nil {
		return errfmt.Format(err)
	}

	// System collections can't be fetched directly
	if collectionID <= 0 {
		fmt.Fprintf(os.Stdout, "System collection: %s (ID: %d)\n", c.Collection, collectionID)

		return nil
	}

	collection, err := client.GetCollection(ctx, collectionID)
	if err != nil {
		return errfmt.Format(err)
	}

	if flags.JSON {
		return output.WriteJSON(os.Stdout, collection)
	}

	output.FormatCollectionDetail(os.Stdout, collection)

	return nil
}

type CollectionsCreateCmd struct {
	Name   string `arg:"" help:"Collection name"`
	Parent string `help:"Parent collection name/ID" short:"p"`
	Color  string `help:"Color hex code (e.g., #ff0000)" short:"c"`
}

func (c *CollectionsCreateCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	req := &api.CreateCollectionRequest{
		Title: c.Name,
		Color: c.Color,
	}

	if c.Parent != "" {
		parentID, resolveErr := client.ResolveCollection(ctx, c.Parent)
		if resolveErr != nil {
			return errfmt.Format(resolveErr)
		}

		req.Parent = &struct {
			ID int `json:"$id"`
		}{ID: parentID}
	}

	collection, err := client.CreateCollection(ctx, req)
	if err != nil {
		return errfmt.Format(err)
	}

	if flags.JSON {
		return output.WriteJSON(os.Stdout, collection)
	}

	fmt.Fprintf(os.Stdout, "Created: %s (ID: %d)\n", collection.Title, collection.ID)

	return nil
}

type CollectionsUpdateCmd struct {
	Collection string `arg:"" help:"Collection name or ID"`
	Name       string `help:"New name" short:"n"`
	Color      string `help:"New color" short:"c"`
}

func (c *CollectionsUpdateCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	collectionID, err := client.ResolveCollection(ctx, c.Collection)
	if err != nil {
		return errfmt.Format(err)
	}

	if collectionID <= 0 {
		return fmt.Errorf("cannot update system collection")
	}

	if c.Name == "" && c.Color == "" {
		return fmt.Errorf("no changes specified; use --name or --color")
	}

	req := &api.UpdateCollectionRequest{
		Title: c.Name,
		Color: c.Color,
	}

	collection, err := client.UpdateCollection(ctx, collectionID, req)
	if err != nil {
		return errfmt.Format(err)
	}

	if flags.JSON {
		return output.WriteJSON(os.Stdout, collection)
	}

	fmt.Fprintf(os.Stdout, "Updated: %s (ID: %d)\n", collection.Title, collection.ID)

	return nil
}

type CollectionsDeleteCmd struct {
	Collection string `arg:"" help:"Collection name or ID"`
}

func (c *CollectionsDeleteCmd) Run(flags *RootFlags) error {
	client, ctx, cancel, err := getClientWithContext()
	if err != nil {
		return errfmt.Format(err)
	}
	defer cancel()

	collectionID, err := client.ResolveCollection(ctx, c.Collection)
	if err != nil {
		return errfmt.Format(err)
	}

	if collectionID <= 0 {
		return fmt.Errorf("cannot delete system collection")
	}

	// Get collection details for confirmation
	collection, err := client.GetCollection(ctx, collectionID)
	if err != nil {
		return errfmt.Format(err)
	}

	msg := fmt.Sprintf("Delete collection '%s' (ID: %d)?", collection.Title, collection.ID)
	if collection.Count > 0 {
		msg = fmt.Sprintf("Delete collection '%s' with %d raindrop(s)?", collection.Title, collection.Count)
	}

	if !confirmAction(msg, flags) {
		fmt.Fprintln(os.Stdout, "Cancelled.")

		return nil
	}

	if err := client.DeleteCollection(ctx, collectionID); err != nil {
		return errfmt.Format(err)
	}

	fmt.Fprintln(os.Stdout, "Deleted.")

	return nil
}

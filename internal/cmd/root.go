package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/alecthomas/kong"

	"github.com/dedene/raindrop-cli/internal/output"
)

type RootFlags struct {
	JSON       bool   `help:"Output JSON to stdout (best for scripting)"`
	Verbose    bool   `help:"Enable verbose logging"`
	Force      bool   `help:"Skip confirmations"`
	NoInput    bool   `help:"Fail instead of prompting (CI mode)" name:"no-input"`
	Hyperlinks string `help:"Hyperlink mode: auto, on, off" default:"auto" enum:"auto,on,off"`
}

// HyperlinkMode returns the parsed hyperlink mode.
func (f *RootFlags) HyperlinkMode() output.HyperlinkMode {
	switch f.Hyperlinks {
	case "on":
		return output.HyperlinkOn
	case "off":
		return output.HyperlinkOff
	default:
		return output.HyperlinkAuto
	}
}

type CLI struct {
	RootFlags `embed:""`

	Version    kong.VersionFlag `help:"Print version and exit"`
	VersionCmd VersionCmd       `cmd:"" name:"version" help:"Print version"`
	Config     ConfigCmd        `cmd:"" help:"Manage configuration"`
	Auth       AuthCmd          `cmd:"" help:"Authentication and credentials"`

	// Core commands
	Add         AddCmd         `cmd:"" help:"Add a bookmark"`
	List        ListCmd        `cmd:"" help:"List bookmarks"`
	Get         GetCmd         `cmd:"" help:"Get bookmark details"`
	Update      UpdateCmd      `cmd:"" help:"Update a bookmark"`
	Delete      DeleteCmd      `cmd:"" help:"Delete a bookmark"`
	Search      SearchCmd      `cmd:"" help:"Search bookmarks"`
	Collections CollectionsCmd `cmd:"" help:"Manage collections"`
	Tags        TagsCmd        `cmd:"" help:"Manage tags"`
	Highlights  HighlightsCmd  `cmd:"" help:"Manage highlights"`

	// Utility commands
	Import     ImportCmd     `cmd:"" help:"Import bookmarks from HTML file"`
	Export     ExportCmd     `cmd:"" help:"Export bookmarks"`
	Open       OpenCmd       `cmd:"" help:"Open bookmark in browser"`
	Copy       CopyCmd       `cmd:"" help:"Copy bookmark URL to clipboard"`
	Completion CompletionCmd `cmd:"" help:"Generate shell completions"`
}

type exitPanic struct{ code int }

func Execute(args []string) (err error) {
	parser, err := newParser()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			if ep, ok := r.(exitPanic); ok {
				if ep.code == 0 {
					err = nil

					return
				}

				err = &ExitError{Code: ep.code, Err: errors.New("exited")}

				return
			}

			panic(r)
		}
	}()

	// Show help when no command provided
	if len(args) == 0 {
		args = []string{"--help"}
	}

	kctx, err := parser.Parse(args)
	if err != nil {
		parsedErr := wrapParseError(err)
		_, _ = fmt.Fprintln(os.Stderr, parsedErr)

		return parsedErr
	}

	err = kctx.Run()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		return err
	}

	return nil
}

func wrapParseError(err error) error {
	if err == nil {
		return nil
	}

	var parseErr *kong.ParseError
	if errors.As(err, &parseErr) {
		return &ExitError{Code: ExitUsage, Err: parseErr}
	}

	return err
}

func newParser() (*kong.Kong, error) {
	vars := kong.Vars{
		"version": VersionString(),
	}

	cli := &CLI{}
	parser, err := kong.New(
		cli,
		kong.Name("raindrop"),
		kong.Description("Raindrop.io CLI - manage bookmarks from the command line"),
		kong.Vars(vars),
		kong.Writers(os.Stdout, os.Stderr),
		kong.Exit(func(code int) { panic(exitPanic{code: code}) }),
		kong.Bind(&cli.RootFlags),
		kong.Help(helpPrinter),
		kong.ConfigureHelp(helpOptions()),
	)
	if err != nil {
		return nil, err
	}

	return parser, nil
}

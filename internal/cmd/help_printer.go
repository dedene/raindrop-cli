package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

const (
	colorAuto  = "auto"
	colorNever = "never"
)

func helpOptions() kong.HelpOptions {
	return kong.HelpOptions{
		NoExpandSubcommands: true,
	}
}

func helpPrinter(options kong.HelpOptions, ctx *kong.Context) error {
	origStdout := ctx.Stdout
	buf := bytes.NewBuffer(nil)
	ctx.Stdout = buf

	defer func() { ctx.Stdout = origStdout }()

	// Set terminal width for proper wrapping
	width := guessColumns(origStdout)
	oldCols, hadCols := os.LookupEnv("COLUMNS")
	_ = os.Setenv("COLUMNS", strconv.Itoa(width))

	defer func() {
		if hadCols {
			_ = os.Setenv("COLUMNS", oldCols)
		} else {
			_ = os.Unsetenv("COLUMNS")
		}
	}()

	if err := kong.DefaultHelpPrinter(options, ctx); err != nil {
		return err
	}

	// Post-process: inject build info and colorize
	out := injectBuildLine(buf.String())
	out = colorizeHelp(out, helpProfile(origStdout, helpColorMode(ctx.Args)))
	_, err := io.WriteString(origStdout, out)

	return err
}

func injectBuildLine(out string) string {
	v := strings.TrimSpace(version)
	if v == "" {
		v = "dev"
	}

	line := fmt.Sprintf("Build: %s", v)
	lines := strings.Split(out, "\n")

	for i, l := range lines {
		if strings.HasPrefix(l, "Usage:") {
			// Don't duplicate if already present
			if i+1 < len(lines) && lines[i+1] == line {
				return out
			}

			result := make([]string, 0, len(lines)+1)
			result = append(result, lines[:i+1]...)
			result = append(result, line)
			result = append(result, lines[i+1:]...)

			return strings.Join(result, "\n")
		}
	}

	return out
}

func helpColorMode(args []string) string {
	if v := os.Getenv("RAINDROP_COLOR"); v != "" {
		return strings.ToLower(strings.TrimSpace(v))
	}

	if slices.Contains(args, "--json") {
		return colorNever
	}

	return colorAuto
}

func helpProfile(stdout io.Writer, mode string) termenv.Profile {
	if termenv.EnvNoColor() {
		return termenv.Ascii
	}

	switch strings.ToLower(strings.TrimSpace(mode)) {
	case colorNever:
		return termenv.Ascii
	case "always":
		return termenv.TrueColor
	default:
		return termenv.NewOutput(stdout, termenv.WithProfile(termenv.EnvColorProfile())).Profile
	}
}

func colorizeHelp(out string, profile termenv.Profile) string {
	if profile == termenv.Ascii {
		return out
	}

	heading := func(s string) string {
		return termenv.String(s).Foreground(profile.Color("#60a5fa")).Bold().String()
	}
	section := func(s string) string {
		return termenv.String(s).Foreground(profile.Color("#a78bfa")).Bold().String()
	}
	cmdName := func(s string) string {
		return termenv.String(s).Foreground(profile.Color("#38bdf8")).Bold().String()
	}
	dim := func(s string) string {
		return termenv.String(s).Foreground(profile.Color("#9ca3af")).String()
	}

	inCommands := false
	lines := strings.Split(out, "\n")

	for i, line := range lines {
		if line == "Commands:" {
			inCommands = true
		}

		switch {
		case strings.HasPrefix(line, "Usage:"):
			lines[i] = heading("Usage:") + strings.TrimPrefix(line, "Usage:")
		case strings.HasPrefix(line, "Build:"):
			lines[i] = section(line)
		case line == "Flags:" || line == "Commands:" || line == "Arguments:":
			lines[i] = section(line)
		case inCommands && strings.HasPrefix(line, "  ") && len(line) > 2 && line[2] != ' ':
			lines[i] = colorizeCommandLine(line, cmdName, dim)
		case inCommands && strings.HasPrefix(line, "    "):
			lines[i] = "    " + dim(strings.TrimPrefix(line, "    "))
		}
	}

	return strings.Join(lines, "\n")
}

func colorizeCommandLine(line string, cmdName, dim func(string) string) string {
	rest := strings.TrimPrefix(line, "  ")
	name, tail, _ := strings.Cut(rest, " ")

	if name == "" {
		return line
	}

	styled := cmdName(name)
	if tail == "" {
		return "  " + styled
	}

	tail = strings.ReplaceAll(tail, "<", dim("<"))
	tail = strings.ReplaceAll(tail, ">", dim(">"))
	tail = strings.ReplaceAll(tail, "[flags]", dim("[flags]"))

	return "  " + styled + " " + tail
}

func guessColumns(w io.Writer) int {
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if n, err := strconv.Atoi(cols); err == nil {
			return n
		}
	}

	if f, ok := w.(*os.File); ok {
		if width, _, err := term.GetSize(int(f.Fd())); err == nil && width > 0 {
			return width
		}
	}

	return 80
}

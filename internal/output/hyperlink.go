package output

import (
	"os"
	"strings"
)

// HyperlinkMode controls hyperlink output.
type HyperlinkMode int

const (
	HyperlinkAuto HyperlinkMode = iota
	HyperlinkOn
	HyperlinkOff
)

// hyperlinkSupported caches the detection result.
var hyperlinkSupported *bool

// SupportsHyperlinks detects OSC 8 support via env heuristics.
func SupportsHyperlinks() bool {
	if hyperlinkSupported != nil {
		return *hyperlinkSupported
	}

	result := detectHyperlinks()
	hyperlinkSupported = &result

	return result
}

func detectHyperlinks() bool {
	// Exclude tmux/screen - no OSC8 passthrough
	term := os.Getenv("TERM")
	if strings.HasPrefix(term, "screen") || strings.HasPrefix(term, "tmux") {
		return false
	}

	if os.Getenv("TMUX") != "" {
		return false
	}

	// Check known supporting terminals
	if prog := os.Getenv("TERM_PROGRAM"); prog != "" {
		switch prog {
		case "iTerm.app", "vscode", "Hyper", "WezTerm", "Tabby":
			return true
		}
	}

	// Kitty
	if os.Getenv("KITTY_WINDOW_ID") != "" {
		return true
	}

	// Ghostty
	if os.Getenv("GHOSTTY_RESOURCES_DIR") != "" {
		return true
	}

	// Alacritty (0.11+)
	if os.Getenv("ALACRITTY_WINDOW_ID") != "" {
		return true
	}

	// Windows Terminal
	if os.Getenv("WT_SESSION") != "" {
		return true
	}

	// VTE-based (Gnome Terminal, etc.) - VTE 0.50+ (version 5000+)
	if vte := os.Getenv("VTE_VERSION"); vte != "" {
		if len(vte) >= 4 && vte[0] >= '5' {
			return true
		}
	}

	return false
}

// Hyperlink wraps text in OSC 8 escape sequence.
func Hyperlink(url, text string) string {
	return output.Hyperlink(url, text)
}

// MaybeHyperlink returns hyperlinked text if mode allows.
func MaybeHyperlink(url, text string, mode HyperlinkMode) string {
	enabled := false

	switch mode {
	case HyperlinkOn:
		enabled = true
	case HyperlinkOff:
		enabled = false
	case HyperlinkAuto:
		enabled = SupportsHyperlinks()
	}

	if enabled {
		return Hyperlink(url, text)
	}

	return text
}

// TruncateURL truncates URL for display.
func TruncateURL(url string, maxWidth int) string {
	if len(url) <= maxWidth {
		return url
	}

	if maxWidth <= 3 {
		return url[:maxWidth]
	}

	return url[:maxWidth-3] + "..."
}

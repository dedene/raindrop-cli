package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/muesli/termenv"
)

// Mode represents the output format mode.
type Mode int

const (
	ModeTable Mode = iota
	ModeJSON
)

var (
	output = termenv.NewOutput(os.Stdout)
	// Color profiles
	colorProfile = output.ColorProfile()
)

// WriteJSON writes the given value as pretty-printed JSON.
func WriteJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	return nil
}

// StyleBold returns a bold styled string.
func StyleBold(s string) string {
	return output.String(s).Bold().String()
}

// StyleFaint returns a faint styled string.
func StyleFaint(s string) string {
	return output.String(s).Faint().String()
}

// StyleGreen returns a green styled string.
func StyleGreen(s string) string {
	return output.String(s).Foreground(colorProfile.Color("2")).String()
}

// StyleRed returns a red styled string.
func StyleRed(s string) string {
	return output.String(s).Foreground(colorProfile.Color("1")).String()
}

// StyleYellow returns a yellow styled string.
func StyleYellow(s string) string {
	return output.String(s).Foreground(colorProfile.Color("3")).String()
}

// StyleBlue returns a blue styled string.
func StyleBlue(s string) string {
	return output.String(s).Foreground(colorProfile.Color("4")).String()
}

// StyleCyan returns a cyan styled string.
func StyleCyan(s string) string {
	return output.String(s).Foreground(colorProfile.Color("6")).String()
}

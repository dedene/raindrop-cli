package output

import (
	"fmt"
	"io"
	"strings"
)

// TableWriter writes formatted tables.
type TableWriter struct {
	w       io.Writer
	headers []string
	rows    [][]string
	widths  []int
}

// NewTableWriter creates a new table writer.
func NewTableWriter(w io.Writer, headers ...string) *TableWriter {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}

	return &TableWriter{
		w:       w,
		headers: headers,
		widths:  widths,
	}
}

// AddRow adds a row to the table.
func (t *TableWriter) AddRow(cols ...string) {
	for i, c := range cols {
		w := VisibleWidth(c)
		if i < len(t.widths) && w > t.widths[i] {
			t.widths[i] = w
		}
	}

	t.rows = append(t.rows, cols)
}

// Render writes the table to the output.
func (t *TableWriter) Render() {
	// Print headers
	for i, h := range t.headers {
		if i > 0 {
			fmt.Fprint(t.w, "  ")
		}

		fmt.Fprint(t.w, StyleBold(pad(h, t.widths[i])))
	}

	fmt.Fprintln(t.w)

	// Print rows
	for _, row := range t.rows {
		for i, col := range row {
			if i > 0 {
				fmt.Fprint(t.w, "  ")
			}

			if i < len(t.widths) {
				fmt.Fprint(t.w, pad(col, t.widths[i]))
			} else {
				fmt.Fprint(t.w, col)
			}
		}

		fmt.Fprintln(t.w)
	}
}

// Count returns the number of rows.
func (t *TableWriter) Count() int {
	return len(t.rows)
}

func pad(s string, width int) string {
	visible := VisibleWidth(s)
	if visible >= width {
		return s
	}

	return s + strings.Repeat(" ", width-visible)
}

// VisibleWidth returns character width ignoring ANSI/OSC escape sequences.
func VisibleWidth(s string) int {
	width := 0
	i := 0

	for i < len(s) {
		if s[i] == '\x1b' && i+1 < len(s) {
			next := s[i+1]
			if next == ']' {
				// OSC sequence: ESC ] ... (ST or BEL)
				// Skip until ST (\x1b\) or BEL (\x07)
				i += 2

				for i < len(s) {
					if s[i] == '\x07' {
						i++

						break
					}

					if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '\\' {
						i += 2

						break
					}

					i++
				}

				continue
			}

			if next == '[' {
				// CSI sequence: ESC [ ... (letter)
				i += 2

				for i < len(s) {
					c := s[i]
					i++

					if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
						break
					}
				}

				continue
			}
		}

		width++
		i++
	}

	return width
}

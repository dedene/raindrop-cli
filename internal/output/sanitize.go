package output

import "strings"

// SanitizeText strips terminal control sequences and non-printable characters.
// It preserves newline and tab for multi-line readable output.
func SanitizeText(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); {
		ch := s[i]

		if ch == '\x1b' {
			if i+1 < len(s) {
				switch s[i+1] {
				case '[':
					// CSI sequence: ESC [ ... final-byte
					i += 2
					for i < len(s) {
						c := s[i]
						i++
						if c >= 0x40 && c <= 0x7e {
							break
						}
					}
					continue
				case ']':
					// OSC sequence: ESC ] ... (BEL or ST)
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
				case 'P', '^', '_':
					// DCS/PM/APC sequence: ESC <byte> ... ST
					i += 2
					for i < len(s) {
						if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '\\' {
							i += 2
							break
						}
						i++
					}
					continue
				default:
					// Other two-byte escape sequence.
					i += 2
					continue
				}
			}

			i++
			continue
		}

		if ch < 0x20 || ch == 0x7f {
			if ch == '\n' || ch == '\t' {
				b.WriteByte(ch)
			}
			i++
			continue
		}

		b.WriteByte(ch)
		i++
	}

	return b.String()
}

// SanitizeInline strips terminal control sequences and line-breaking controls.
func SanitizeInline(s string) string {
	safe := SanitizeText(s)
	safe = strings.ReplaceAll(safe, "\n", " ")
	safe = strings.ReplaceAll(safe, "\t", " ")
	return strings.TrimSpace(safe)
}

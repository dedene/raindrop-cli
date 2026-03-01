package output

import "testing"

func TestSanitizeTextStripsControlSequences(t *testing.T) {
	raw := "hello\x1b[31m red\x1b[0m world\x1b]0;title\x07!\x00"
	got := SanitizeText(raw)
	want := "hello red world!"

	if got != want {
		t.Fatalf("unexpected sanitized text:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestSanitizeTextPreservesNewlinesAndTabs(t *testing.T) {
	raw := "line1\nline2\tvalue\r"
	got := SanitizeText(raw)
	want := "line1\nline2\tvalue"

	if got != want {
		t.Fatalf("unexpected sanitized text:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestSanitizeInlineRemovesLineBreaks(t *testing.T) {
	raw := "title\nwith\tspaces"
	got := SanitizeInline(raw)
	want := "title with spaces"

	if got != want {
		t.Fatalf("unexpected inline sanitized text:\nwant: %q\ngot:  %q", want, got)
	}
}

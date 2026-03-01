package browser

import "testing"

func TestOpenCommandWindowsUsesRundll32(t *testing.T) {
	url := "https://example.com/?q=test"

	cmd, err := openCommand("windows", url)
	if err != nil {
		t.Fatalf("openCommand returned error: %v", err)
	}

	if len(cmd.Args) < 3 {
		t.Fatalf("expected at least 3 args, got %v", cmd.Args)
	}

	if cmd.Args[0] != "rundll32" {
		t.Fatalf("expected rundll32 launcher, got %q", cmd.Args[0])
	}

	if cmd.Args[1] != "url.dll,FileProtocolHandler" {
		t.Fatalf("unexpected windows handler arg: %q", cmd.Args[1])
	}

	if cmd.Args[2] != url {
		t.Fatalf("unexpected url arg: %q", cmd.Args[2])
	}
}

func TestOpenCommandUnsupportedPlatform(t *testing.T) {
	if _, err := openCommand("plan9", "https://example.com"); err == nil {
		t.Fatal("expected error for unsupported platform")
	}
}

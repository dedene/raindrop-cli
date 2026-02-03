package main

import (
	"os"

	"github.com/dedene/raindrop-cli/internal/cmd"
)

func main() {
	err := cmd.Execute(os.Args[1:])
	os.Exit(cmd.ExitCode(err))
}

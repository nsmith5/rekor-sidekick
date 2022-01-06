package main

import (
	// Loading output drivers
	_ "github.com/nsmith5/rekor-sidekick/outputs"
	_ "github.com/nsmith5/rekor-sidekick/outputs/stdout"
)

func main() {
	cmd := newCLI()
	cmd.Execute()
}

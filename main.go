package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// ErrNotImplimented signals incomplete code path
	ErrNotImplimented = errors.New(`rekor-sidekick: not implemented`)
)

func newCLI() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rekor-sidekick",
		Short: "Transparency log monitoring and alerting",
		Long:  "Daemon that monitors a Rekor instance and forwards entries of interest to configurable destinations",
		Run:   runCLI,
	}

	return cmd
}

func runCLI(cmd *cobra.Command, args []string) {
	fmt.Println(ErrNotImplimented)
	os.Exit(1)
}

func main() {
	cmd := newCLI()

	cmd.Execute()
}

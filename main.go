package main

import (
	"log"

	"github.com/nsmith5/rekor-sidekick/cli"

	// Loading output drivers
	_ "github.com/nsmith5/rekor-sidekick/outputs"
	_ "github.com/nsmith5/rekor-sidekick/outputs/cloudevents"
	_ "github.com/nsmith5/rekor-sidekick/outputs/opensearch"
	_ "github.com/nsmith5/rekor-sidekick/outputs/pagerduty"
	_ "github.com/nsmith5/rekor-sidekick/outputs/stdout"
)

func main() {
	cmd := cli.New()
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

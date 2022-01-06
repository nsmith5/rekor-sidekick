package stdout

import (
	"fmt"

	"github.com/nsmith5/rekor-sidekick/outputs"
)

type impl struct{}

func (i impl) Send(e outputs.Event) error {
	fmt.Printf(
		`{"name": "%s", "description": "%s", rekorURL: "%s"}`,
		e.Name,
		e.Description,
		e.RekorURL,
	)
	fmt.Println()
	return nil
}

func New(map[string]interface{}) (outputs.Output, error) {
	return &impl{}, nil
}

func init() {
	outputs.RegisterDriver("stdout", outputs.CreatorFunc(New))
}

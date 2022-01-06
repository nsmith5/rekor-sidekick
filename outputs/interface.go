package outputs

import (
	"github.com/nsmith5/rekor-sidekick/policy"
	"github.com/nsmith5/rekor-sidekick/rekor"
)

type Event struct {
	Policy policy.Policy
	Entry  rekor.LogEntry
}

type Output interface {
	Send(Event) error
}

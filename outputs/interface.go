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
	// Send pushes an alert event to a driver specific backend
	Send(Event) error

	// Name returns driver name
	Name() string
}

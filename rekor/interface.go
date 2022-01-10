package rekor

import (
	"errors"
	"github.com/sigstore/rekor/pkg/generated/models"
)

var (
	// ErrEntryDoesntExist signals a log entry that hasn't made it into the Rekor log just yet
	ErrEntryDoesntExist = errors.New(`Rekor entry doesn't exist yet`)
)

// Client is a Rekor API client
type Client interface {
	// GetEntry pulls a specificy rekor log entry by index.
	GetEntry(uint) (models.LogEntry, error)

	// GetNextEntry pulls the next entry in the Rekor log. If the
	// next log doesn't exist yet ErrEntryDoesntExist is returned.
	GetNextEntry() (models.LogEntry, error)

	// GetTreeState fetches the current state of the rekor log including
	// log size
	GetTreeState() (*models.LogInfo, error)
}

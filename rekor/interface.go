package rekor

import (
	"errors"
	"time"
)

var (
	// ErrEntryDoesntExist signals a log entry that hasn't made it into the Rekor log just yet
	ErrEntryDoesntExist = errors.New(`Rekor entry doesn't exist yet`)
)

// LogEntry is a Rekor log entry
type LogEntry struct {
	URL          string
	IntegratedAt time.Time
	Index        uint
	Body         map[string]interface{}
}

// TreeState represents the current state of the transparency log (size
// etc)
type TreeState struct {
	RootHash       string
	SignedTreeHead string
	TreeSize       uint
}

// Client is a Rekor API client
type Client interface {
	// GetEntry pulls a specificy rekor log entry by index.
	GetEntry(uint) (*LogEntry, error)

	// GetNextEntry pulls the next entry in the Rekor log. If the
	// next log doesn't exist yet ErrEntryDoesntExist is returned.
	GetNextEntry() (*LogEntry, error)

	// GetTreeState fetches the current state of the rekor log including
	// log size
	GetTreeState() (*TreeState, error)
}

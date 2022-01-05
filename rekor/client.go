package rekor

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	// ErrEntryDoesntExist signals a log entry that hasn't made it into the Rekor log just yet
	ErrEntryDoesntExist = errors.New(`Rekor entry doesn't exist yet`)
)

// LogEntry is a Rekor log entry
type LogEntry map[string]interface{}

// treeState represents the current state of the transparency log (size
// etc)
type treeState struct {
	RootHash       string
	SignedTreeHead string
	TreeSize       uint
}

// Client is a Rekor api client
type Client struct {
	baseURL      string
	currentIndex uint

	*http.Client
}

// NewClient returns a Rekor client or fails if the baseURL
// is misconfigured.
func NewClient(baseURL string) (*Client, error) {
	rc := Client{
		baseURL:      baseURL,
		currentIndex: 0,
		Client:       new(http.Client),
	}

	// Grab the latest signed tree state and use the tree size as a starting
	// point to start iterating log entries. Its not the very tip of the log,
	// but its close enough for us.
	state, err := rc.getTreeState()
	if err != nil {
		// If this bailed... we're going to guess its probably misconfiguration
		// not a temporary outage. Lets just bail hard.
		return nil, fmt.Errorf("failed to get initial tree state. Is rekor server configured correctly? Failured caused by %w", err)
	}
	rc.currentIndex = state.TreeSize

	return &rc, nil
}

func (rc *Client) getLogEntry(index uint) (LogEntry, error) {
	url := fmt.Sprintf("%s/api/v1/log/entries?logIndex=%d", rc.baseURL, index)

	req, err := http.NewRequest(`GET`, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(`Accept`, `application/json`)
	resp, err := rc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrEntryDoesntExist
	}

	m := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&m)
	if err != nil {
		return nil, err
	}

	var body string
	// The key is entry UUID, but we have no idea what that is apriori so we
	// grab it by looping over key-value pairs and breaking after one
	for _, v := range m {
		m, ok := v.(map[string]interface{})
		if !ok {
			return nil, errors.New(`malformed rekor entry response`)
		}
		body, ok = m["body"].(string)
		if !ok {
			return nil, errors.New(`malformed rekor entry response`)
		}
		break
	}

	var entry = make(LogEntry)
	err = json.NewDecoder(
		base64.NewDecoder(base64.URLEncoding, strings.NewReader(body)),
	).Decode(&entry)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// GetNextLogEntry pulls the next entry in the Rekor log. If the
// next log doesn't exist yet ErrEntryDoesntExist is returned.
func (rc *Client) GetNextLogEntry() (LogEntry, error) {
	entry, err := rc.getLogEntry(rc.currentIndex)
	if err != nil {
		return nil, err
	}
	rc.currentIndex++
	return entry, nil
}

func (rc *Client) getTreeState() (*treeState, error) {
	url := fmt.Sprintf("%s/api/v1/log", rc.baseURL)

	req, err := http.NewRequest(`GET`, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(`Accept`, `application/json`)
	resp, err := rc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var state treeState
	err = json.NewDecoder(resp.Body).Decode(&state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}

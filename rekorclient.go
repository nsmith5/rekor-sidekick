package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// rekorEntry is unstructured, we create this type simply
// to know what we're talking about passing it around
type rekorLogEntry struct {
	Kind       string
	APIVersion string `json:"apiVersion"`
	Spec       interface{}
}

// rekorTreeState represents the current state of the transparency log (size
// etc)
type rekorTreeState struct {
	RootHash       string
	SignedTreeHead string
	TreeSize       uint
}

type rekorClient struct {
	baseURL      string
	currentIndex uint

	*http.Client
}

func newRekorClient(baseURL string) (*rekorClient, error) {
	rc := rekorClient{
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

func (rc *rekorClient) getLogEntry(index uint) (*rekorLogEntry, error) {
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

	var entry rekorLogEntry
	err = json.NewDecoder(
		base64.NewDecoder(base64.URLEncoding, strings.NewReader(body)),
	).Decode(&entry)
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (rc *rekorClient) getNextLogEntry() (*rekorLogEntry, error) {
	entry, err := rc.getLogEntry(rc.currentIndex)
	if err != nil {
		return nil, err
	}
	rc.currentIndex++
	return entry, nil
}

func (rc *rekorClient) getTreeState() (*rekorTreeState, error) {
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

	var state rekorTreeState
	err = json.NewDecoder(resp.Body).Decode(&state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}

package rekor

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type impl struct {
	baseURL      string
	currentIndex uint

	*http.Client
}

// NewClient returns a Rekor client or fails if the baseURL
// is misconfigured.
func NewClient(baseURL string, index int) (Client, error) {
	rc := impl{
		baseURL:      baseURL,
		currentIndex: 0,
		Client:       new(http.Client),
	}

	// No starting index provided by the config
	if index == -1 {
	// Grab the latest signed tree state and use the tree size as a starting
	// point to start iterating log entries. Its not the very tip of the log,
	// but its close enough for us.
	state, err := rc.GetTreeState()
	if err != nil {
		// If this bailed... we're going to guess its probably misconfiguration
		// not a temporary outage. Lets just bail hard.
		return nil, fmt.Errorf("failed to get initial tree state. Is rekor server configured correctly? Failured caused by %w", err)
	}
	rc.currentIndex = state.TreeSize
	} else {
		rc.currentIndex = uint(index)
	}

	return &rc, nil
}

func (rc *impl) GetEntry(index uint) (*LogEntry, error) {
	var entry LogEntry

	entry.Index = index

	entryMap := make(map[string]interface{})
	{
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

		err = json.NewDecoder(resp.Body).Decode(&entryMap)
		if err != nil {
			return nil, err
		}
	}

	var (
		uuid string
		unix float64
		body string
	)
	{
		// The key is entry UUID, but we have no idea what that is apriori so we
		// grab it by looping over key-value pairs and breaking after one
		for id, v := range entryMap {
			uuid = id

			m, ok := v.(map[string]interface{})
			if !ok {
				return nil, errors.New(`malformed rekor entry response`)
			}

			unix, ok = m[`integratedTime`].(float64)
			if !ok {
				return nil, errors.New(`malformed rekor integration time`)
			}

			body, ok = m["body"].(string)
			if !ok {
				return nil, errors.New(`malformed rekor entry response`)
			}
			break
		}
	}

	// (1) UUID -> URL
	entry.URL = fmt.Sprintf("%s/api/v1/log/entries/%s", rc.baseURL, uuid)

	// (2) Unix time -> created time
	entry.IntegratedAt = time.Unix(int64(unix), 0)

	// (3) Decode body
	decodedBody := make(map[string]interface{})
	err := json.NewDecoder(
		base64.NewDecoder(base64.URLEncoding, strings.NewReader(body)),
	).Decode(&decodedBody)
	if err != nil {
		return nil, err
	}
	entry.Body = decodedBody

	return &entry, nil
}

func (rc *impl) GetNextEntry() (*LogEntry, error) {
	entry, err := rc.GetEntry(rc.currentIndex)
	if err != nil {
		return nil, err
	}
	rc.currentIndex++
	return entry, nil
}

func (rc *impl) GetTreeState() (*TreeState, error) {
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

	var state TreeState
	err = json.NewDecoder(resp.Body).Decode(&state)
	if err != nil {
		return nil, err
	}

	return &state, nil
}

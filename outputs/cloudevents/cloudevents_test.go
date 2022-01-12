package cloudevents

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nsmith5/rekor-sidekick/outputs"
)

func TestCreateDriver(t *testing.T) {
	tests := map[string]struct {
		Conf  map[string]interface{}
		Error error
	}{
		"valid configuration": {
			Conf: map[string]interface{}{
				`sourceID`: `id`,
				`http`: map[string]interface{}{
					`url`: `https://localhost:8080`,
				},
			},
			Error: nil,
		},
		"missing url in config": {
			Conf: map[string]interface{}{
				`sourceID`: `abc123`,
			},
			Error: ErrConfigMissingURL,
		},
		"missing sourceID in config": {
			Conf: map[string]interface{}{
				`http`: map[string]interface{}{
					`url`: `http://derp:8080`,
				},
			},
			Error: ErrConfigMissingSourceID,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := createDriver(data.Conf)
			if err != data.Error {
				t.Errorf("Expected err %q, but recieved %q", data.Error, err)
			}
		})
	}
}

func TestSend(t *testing.T) {
	sourceID := `23124131`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedSource := fmt.Sprintf("%s:%s", eventSourcePrefix, sourceID)
		if r.Header.Get("Ce-Source") != expectedSource {
			t.Errorf("expected event source to be %s but got %s", expectedSource, r.Header.Get("Ce-Source"))
		}
		if r.Header.Get("Ce-Type") != eventType {
			t.Errorf("expected event type to be %s but got %s", eventType, r.Header.Get("Ce-Type"))
		}
	}))

	conf := map[string]interface{}{
		`sourceID`: sourceID,
		`http`: map[string]interface{}{
			`url`: ts.URL,
		},
	}

	driver, err := createDriver(conf)
	if err != nil {
		t.Fatal(err)
	}

	err = driver.Send(outputs.Event{})
	if err != nil {
		t.Error(err)
	}
}

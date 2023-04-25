package opensearch

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/nsmith5/rekor-sidekick/outputs"
	"github.com/nsmith5/rekor-sidekick/rekor"

	opensearch "github.com/opensearch-project/opensearch-go"
	opensearchapi "github.com/opensearch-project/opensearch-go/opensearchapi"
)

const (
	driverName = `opensearch`
)

type config struct {
	Severity string
	Server   string
	Insecure bool
	Index    string
	Username string
	Password string
}

type driver struct {
	index string

	client *opensearch.Client
}

// Flatten the event and create a reader
// ALSO returns the id (URL) as a string
func Searchify(e rekor.LogEntry) (*strings.Reader, string, error) {
	// Struct -> map[string]interface{}
	data, err := json.Marshal(e)
	if err != nil {
		return nil, "", err
	}

	var entry map[string]interface{}
	json.Unmarshal(data, &entry)

	// flatten the map
	entry = Flatten(entry)
	url := entry["URL"].(string)
	slice := strings.Split(url, "/")
	entryId := slice[len(slice)-1]

	b, err := json.Marshal(entry)
	if err != nil {
		return nil, "", err
	}

	return strings.NewReader(string(b)), entryId, nil
}

// https://stackoverflow.com/a/39625223
// Flatten takes a map and returns a new one where nested maps are replaced
// by dot-delimited keys.
func Flatten(m map[string]interface{}) map[string]interface{} {
	o := make(map[string]interface{})
	for k, v := range m {
		switch child := v.(type) {
		case map[string]interface{}:
			nm := Flatten(child)
			for nk, nv := range nm {
				o[k+"."+nk] = nv
			}
		default:
			o[k] = v
		}
	}
	return o
}

func (d *driver) Send(e outputs.Event) error {

	document, id, err := Searchify(e.Entry)
	if err != nil {
		return err
	}

	req := opensearchapi.IndexRequest{
		Index:      d.index,
		DocumentID: id,
		Body:       document,
	}
	res, err := req.Do(context.Background(), d.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

func (d *driver) Name() string {
	return driverName
}

func createDriver(conf map[string]interface{}) (outputs.Output, error) {
	var c config
	err := mapstructure.Decode(conf, &c)
	if err != nil {
		return nil, err
	}

	if c.Server == "" {
		return nil, errors.New(`opensearch: server url required (e.g. https://localhost:9200)`)
	}
	if c.Index == "" {
		return nil, errors.New(`opensearch: index required (will be created if doesn't exist)`)
	}
	if c.Username == "" {
		return nil, errors.New(`opensearch: username required`)
	}
	if c.Password == "" {
		return nil, errors.New(`opensearch: password required`)
	}

	// Optional insecure flag
	var transport *http.Transport
	if c.Insecure {
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	client, err := opensearch.NewClient(opensearch.Config{
		Transport: transport,
		Addresses: []string{c.Server},
		Username:  c.Username,
		Password:  c.Password,
	})

	if err != nil {
		return nil, fmt.Errorf("opensearch: failed to create client: %s", err.Error())
	}

	return &driver{
		index:  c.Index,
		client: client,
	}, nil
}

func init() {
	outputs.RegisterDriver(driverName, outputs.CreatorFunc(createDriver))
}

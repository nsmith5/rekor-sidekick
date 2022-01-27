package cloudevents

import (
	"context"
	"errors"
	"fmt"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/mitchellh/mapstructure"
	"github.com/nsmith5/rekor-sidekick/outputs"
)

const (
	driverName        = `cloudevents`
	eventSourcePrefix = `github.com/nsmith5/rekor-sidekick`
	eventType         = `rekor-sidekick.policy.violation.v1`
)

var (
	ErrConfigMissingURL      = errors.New(`cloudevents: driver requires "http.url" in configuration`)
	ErrConfigMissingSourceID = errors.New(`cloudevents: driver requires "sourceID" in configuration`)
)

type config struct {
	SourceID string
	HTTP     struct {
		URL string
	}
}

type driver struct {
	sourceID string
	http     struct {
		url string
	}
	client client.Client
}

func (d *driver) Send(e outputs.Event) error {
	event := ce.NewEvent()
	event.SetSource(fmt.Sprintf("%s:%s", eventSourcePrefix, d.sourceID))
	event.SetType(eventType)
	err := event.SetData(ce.ApplicationJSON, e)
	if err != nil {
		return err
	}

	ctx := ce.ContextWithTarget(context.Background(), d.http.url)

	if result := d.client.Send(ctx, event); !ce.IsACK(result) {
		return result
	}

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

	if c.SourceID == "" {
		return nil, ErrConfigMissingSourceID
	}
	if c.HTTP.URL == "" {
		return nil, ErrConfigMissingURL
	}

	client, err := ce.NewClientHTTP()
	if err != nil {
		return nil, err
	}

	return &driver{
		sourceID: c.SourceID,
		http: struct{ url string }{
			url: c.HTTP.URL,
		},
		client: client,
	}, nil
}

func init() {
	outputs.RegisterDriver(driverName, outputs.CreatorFunc(createDriver))
}

package pagerduty

import (
	"context"
	"errors"

	"github.com/mitchellh/mapstructure"
	"github.com/nsmith5/rekor-sidekick/outputs"

	pd "github.com/PagerDuty/go-pagerduty"
)

const (
	driverName = `pagerduty`
)

type config struct {
	Severity       string
	IntegrationKey string
	APIToken       string
}

type driver struct {
	severity       string
	integrationKey string

	client *pd.Client
}

func (d *driver) Send(e outputs.Event) error {
	payload := pd.V2Payload{
		Summary:  e.Policy.Description,
		Source:   e.Entry.URL,
		Severity: d.severity,
		Group:    e.Policy.Name,
		Class:    `rekor-sidekick.policy.violation.v1`,
	}

	event := pd.V2Event{
		RoutingKey: d.integrationKey,
		Action:     `trigger`,
		Images:     nil,
		Client:     `rekor-sidekick`,
		Payload:    &payload,
	}

	_, err := d.client.ManageEventWithContext(context.Background(), &event)
	if err != nil {
		return err
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

	if c.APIToken == "" {
		return nil, errors.New(`pagerduty: API token required`)
	}
	client := pd.NewClient(c.APIToken)

	return &driver{
		severity:       c.Severity,
		integrationKey: c.IntegrationKey,
		client:         client,
	}, nil
}

func init() {
	outputs.RegisterDriver(driverName, outputs.CreatorFunc(createDriver))
}

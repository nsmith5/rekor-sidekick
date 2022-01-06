package policy

import (
	"context"

	"github.com/open-policy-agent/opa/rego"
)

// Policy specifies when to alert on a Rekor log entry
type Policy struct {
	Name        string
	Description string
	Body        string
}

// Alert checks one log entry and returns true of we should alert on the entry
func (p Policy) Alert(logEntry map[string]interface{}) (bool, error) {
	r := rego.New(
		rego.Query("data.sidekick.alert"),
		rego.Module(p.Name, p.Body),
		rego.Input(logEntry),
	)
	rs, err := r.Eval(context.Background())
	if err != nil {
		return false, err
	}
	return rs.Allowed(), nil
}

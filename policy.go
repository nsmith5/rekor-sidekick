package main

import (
	"context"

	"github.com/nsmith5/rekor-sidekick/rekor"
	"github.com/open-policy-agent/opa/rego"
)

type policy struct {
	Name        string
	Description string
	Body        string
}

func (p policy) allowed(e rekor.LogEntry) (bool, error) {
	r := rego.New(
		rego.Query("data.auth.allow"),
		rego.Module(p.Name, p.Body),
		rego.Input(e),
	)
	rs, err := r.Eval(context.Background())
	if err != nil {
		return false, err
	}
	return rs.Allowed(), nil
}

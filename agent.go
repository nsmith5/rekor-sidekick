package main

import "context"

type agent struct {
}

// newAgent constructs an agent from config or bails
func newAgent(c config) (*agent, error) {
	return nil, ErrNotImplimented
}

// run starts off the agent. The call blocks or exits returning an error
// if the agent hits a fatal error.
func (a *agent) run() error {
	return ErrNotImplimented
}

// shutdown gracefully stops the agent. Shutdown can take an arbitrarily long time. Use
// context cancellation to force shutdown.
func (a *agent) shutdown(context.Context) error {
	return ErrNotImplimented
}

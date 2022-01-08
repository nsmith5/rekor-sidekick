package agent

import "context"

// Agent is a daemon that, while running, pulls rekor entries and checks them
// against a set of alerting policies.
type Agent interface {
	// Run starts off the agent. The call blocks or exits returning an error
	// if the agent hits a fatal error.
	Run() error

	// shutdown gracefully stops the agent. Shutdown can take an arbitrarily long time. Use
	// context cancellation to force shutdown. Calling shutdown more than once will cause a
	// panic.
	Shutdown(context.Context) error
}

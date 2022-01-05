package main

import (
	"context"
	"errors"
	"time"

	"github.com/nsmith5/rekor-sidekick/outputs"
	"github.com/nsmith5/rekor-sidekick/rekor"
)

type agent struct {
	rc       *rekor.Client
	policies []policy
	outs     []outputs.Output

	quit chan struct{}
}

// newAgent constructs an agent from config or bails
func newAgent(c config) (*agent, error) {
	rc, err := rekor.NewClient(c.RekorServerURL)
	if err != nil {
		return nil, err
	}

	policies := c.Policies

	var outs []outputs.Output
	for name, conf := range c.Outputs {
		output, err := outputs.LoadDriver(name, conf)
		if err != nil {
			// Huh... log this issue I guess?
			continue
		}
		outs = append(outs, output)
	}

	quit := make(chan struct{})

	return &agent{rc, policies, outs, quit}, nil
}

// run starts off the agent. The call blocks or exits returning an error
// if the agent hits a fatal error.
func (a *agent) run() error {
	const initialBackoff = time.Duration(10)
	var currentBackoff = time.Duration(10)

	for {
		select {
		case _, ok := <-a.quit:
			if !ok {
				// Should be unreachable as we're supposed to close the channel
				// ourselves!
				return errors.New(`agent: quit chan closed. agent state corrupted`)
			}

			// Close channel to signal to shutdown caller we've cleanly shutdown
			close(a.quit)

			return nil

		default:
			entry, err := a.rc.GetNextLogEntry()
			if err != nil {
				if err == rekor.ErrEntryDoesntExist {
					// Log doesn't exist yet, lets just wait 10 seconds and try again
					time.Sleep(10 * time.Second)

				} else {
					// Lets assume a temporary outage and retry with exponential backoff
					time.Sleep(currentBackoff * time.Second)
					currentBackoff *= 2
				}
				break
			}

			// Incase we just recovered from a temporary outage, lets reset the backoff
			currentBackoff = initialBackoff

			// Policy checks!
			for _, p := range a.policies {
				violation, err := p.allowed(entry)
				if err != nil {
					// huh... what to do here?
					continue
				}

				if violation {
					for _, out := range a.outs {
						// TODO: Populate the rekor URL!
						e := outputs.Event{
							Name:        p.Name,
							Description: p.Description,
							RekorURL:    `dunno...`,
						}

						// TODO: Do something on send failure
						out.Send(e)
					}
				}
			}
		}
	}
}

// shutdown gracefully stops the agent. Shutdown can take an arbitrarily long time. Use
// context cancellation to force shutdown. Calling shutdown more than once will cause a
// panic.
func (a *agent) shutdown(ctx context.Context) error {
	a.quit <- struct{}{}

	select {
	case <-a.quit:
		// Graceful shutdown complete
		return nil

	case <-ctx.Done():
		// We took too long shutting down and the caller is
		// angry. Time to give up
		return errors.New(`timeout on graceful shutdown of agent`)
	}
}

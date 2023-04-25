package agent

import (
	"context"
	"errors"
	"time"

	"github.com/nsmith5/rekor-sidekick/outputs"
	"github.com/nsmith5/rekor-sidekick/policy"
	"github.com/nsmith5/rekor-sidekick/rekor"
)

type impl struct {
	rc       rekor.Client
	policies []policy.Policy
	outs     []outputs.Output

	quit chan struct{}
}

// New constructs an agent from config or bails
func New(c Config) (Agent, error) {
	log := newLogger(c)

	rc, err := rekor.NewClient(c.Server, c.Index)
	if err != nil {
		return nil, err
	}
	rc = rekor.WithLogging(rc, log)

	policies := c.Policies

	var outs []outputs.Output
	for name, conf := range c.Outputs {
		output, err := outputs.LoadDriver(name, conf)
		if err != nil {
			continue
		}
		outs = append(outs, outputs.WithLogging(output, log))
	}

	if len(outs) == 0 {
		return nil, errors.New(`zero output drivers configured`)
	}

	quit := make(chan struct{})

	return withLogging(&impl{rc, policies, outs, quit}, log), nil
}

func (a *impl) Run() error {
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
			entry, err := a.rc.GetNextEntry()
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
				alert, err := p.Alert(entry.Body)
				if err != nil {
					continue
				}

				if alert {
					for _, out := range a.outs {
						// TODO(nsmith5): retry on error?
						_ = out.Send(outputs.Event{Policy: p, Entry: *entry})
					}
				}
			}
		}
	}
}

func (a *impl) Shutdown(ctx context.Context) error {
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

package agent

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/nsmith5/rekor-sidekick/outputs"
	"github.com/nsmith5/rekor-sidekick/policy"
	"github.com/nsmith5/rekor-sidekick/rekor"
	logrus "github.com/sirupsen/logrus"
)

type impl struct {
	rc       *rekor.Client
	policies []policy.Policy
	outs     []outputs.Output

	log *logrus.Logger

	quit chan struct{}
}

// New constructs an agent from config or bails
func New(c Config) (Agent, error) {
	log := logrus.New()
	log.SetOutput(os.Stdout)
	switch c.Logging.Level {
	case "panic":
		log.SetLevel(logrus.PanicLevel)
	case "fatal":
		log.SetLevel(logrus.FatalLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "trace":
		log.SetLevel(logrus.TraceLevel)
	}
	log.SetFormatter(&logrus.JSONFormatter{})
	log.SetReportCaller(true)

	rc, err := rekor.NewClient(c.Server)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err": err,
		}).Error(`failed to create rekor client when creating client`)
		return nil, err
	}

	policies := c.Policies

	var outs []outputs.Output
	for name, conf := range c.Outputs {
		output, err := outputs.LoadDriver(name, conf)
		if err != nil {
			log.WithFields(logrus.Fields{
				"err": err,
			}).Errorf("failed to load driver %s", name)
			continue
		}
		log.Infof("Loaded output driver %s", name)
		outs = append(outs, output)
	}

	if len(outs) == 0 {
		log.Errorf("zero output drivers configured")
		return nil, errors.New(`zero output drivers configured`)
	}

	quit := make(chan struct{})

	return &impl{rc, policies, outs, log, quit}, nil
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
			entry, err := a.rc.GetNextLogEntry()
			if err != nil {
				if err == rekor.ErrEntryDoesntExist {
					// Log doesn't exist yet, lets just wait 10 seconds and try again
					a.log.Debug("no more log entries. sleeping before retry")
					time.Sleep(10 * time.Second)

				} else {
					// Lets assume a temporary outage and retry with exponential backoff
					a.log.WithFields(logrus.Fields{
						`err`:     err,
						`backoff`: currentBackoff,
					}).Errorf("error pulling log entry. retrying with exponential backoff")
					time.Sleep(currentBackoff * time.Second)
					currentBackoff *= 2
				}
				break
			}

			a.log.Debug(`pulled an entry`)

			// Incase we just recovered from a temporary outage, lets reset the backoff
			currentBackoff = initialBackoff

			// Policy checks!
			for _, p := range a.policies {
				a.log.Tracef("checking policy %s", p.Name)

				alert, err := p.Alert(entry.Body)
				if err != nil {
					a.log.WithFields(logrus.Fields{
						`err`: err,
					}).Errorf("failure to evalute policy %s against entry", p.Name)
					continue
				}

				if alert {
					a.log.Debugf("alerting on policy %s", p.Name)
					for _, out := range a.outs {
						err = out.Send(outputs.Event{Policy: p, Entry: *entry})
						if err != nil {
							a.log.WithFields(logrus.Fields{
								`err`: err,
							}).Error("failed to send policy alert event")
						} else {
							a.log.Debug("sent policy alert event")
						}
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

package outputs

import "github.com/sirupsen/logrus"

type ologger struct {
	inner Output
	log   *logrus.Logger
}

// WithLogging logs an output driver
func WithLogging(o Output, log *logrus.Logger) Output {
	return &ologger{inner: o, log: log}
}

func (ol *ologger) Send(e Event) error {
	ol.log.Debugf("output: driver %s sending event", ol.inner.Name())
	err := ol.inner.Send(e)
	if err != nil {
		ol.log.WithFields(logrus.Fields{
			`err`:    err,
			`driver`: ol.inner.Name(),
		}).Error("output: failed to send event")
		return err
	}
	return nil
}

func (ol *ologger) Name() string {
	return ol.inner.Name()
}

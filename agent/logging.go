package agent

import (
	"context"
	"os"

	logrus "github.com/sirupsen/logrus"
)

func newLogger(c Config) *logrus.Logger {
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

	return log
}

type logAgent struct {
	inner Agent
	log   *logrus.Logger
}

// WithLogging adds logging to agent
func withLogging(a Agent, log *logrus.Logger) Agent {
	return &logAgent{inner: a, log: log}
}

func (a *logAgent) Run() error {
	a.log.Info("agent: run launching")
	err := a.inner.Run()
	if err != nil {
		a.log.WithFields(logrus.Fields{
			`err`: err,
		}).Error("agent: exit with error")
		return err
	}
	a.log.Debug("agent: run exit without error")
	return nil
}

func (a *logAgent) Shutdown(ctx context.Context) error {
	a.log.Info("agent: shutting down")
	err := a.inner.Shutdown(ctx)
	if err != nil {
		a.log.WithFields(logrus.Fields{
			`err`: err,
		}).Error("agent: failed to shutdown gracefully")
		return err
	}
	a.log.Debug("agent: successful graceful shutdown")
	return nil
}

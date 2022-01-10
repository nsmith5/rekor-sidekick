package rekor

import (
	"github.com/sigstore/rekor/pkg/generated/models"
	"github.com/sirupsen/logrus"
)

type logClient struct {
	inner Client
	log   *logrus.Logger
}

// WithLogging adds logging to a Rekor client
func WithLogging(c Client, log *logrus.Logger) Client {
	return &logClient{inner: c, log: log}
}

func (rc *logClient) GetEntry(index uint) (models.LogEntry, error) {
	entry, err := rc.inner.GetEntry(index)
	if err != nil {
		rc.log.WithFields(logrus.Fields{
			`err`: err,
		}).Errorf("Failed to fetch log entry %d", index)
		return nil, err
	}
	return entry, nil
}

func (rc *logClient) GetNextEntry() (models.LogEntry, error) {
	return rc.inner.GetNextEntry()
}

func (rc *logClient) GetTreeState() (*models.LogInfo, error) {
	return rc.inner.GetTreeState()
}

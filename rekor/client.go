package rekor

import (
	"fmt"
	rekor "github.com/sigstore/rekor/pkg/client"
	rekorclient "github.com/sigstore/rekor/pkg/generated/client"
	"github.com/sigstore/rekor/pkg/generated/client/entries"
	"github.com/sigstore/rekor/pkg/generated/client/tlog"
	"github.com/sigstore/rekor/pkg/generated/models"
)

type impl struct {
	baseURL      string
	currentIndex uint

	rekorClient *rekorclient.Rekor
}

// TODO: once we provide a version information about the project, we can use this information in userAgent header
//var (
//	// uaString is meant to resemble the User-Agent sent by browsers with requests.
//	// See: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/User-Agent
//	uaString = fmt.Sprintf("rekor-sidekick/%s (%s; %s)", version.GitVersion, runtime.GOOS, runtime.GOARCH)
//)

// NewClient returns a Rekor client or fails if the baseURL
// is misconfigured.
func NewClient(rekorURL string) (Client, error) {
	//rekorClient, err := rekor.GetRekorClient(rekorURL, rekor.WithUserAgent(options.UserAgent()))
	rekorClient, err := rekor.GetRekorClient(rekorURL)
	if err != nil {
		return nil, err
	}

	rc := impl{
		baseURL:      rekorURL,
		currentIndex: 0,
		rekorClient:  rekorClient,
	}

	// Grab the latest signed tree state and use the tree size as a starting
	// point to start iterating log entries. Its not the very tip of the log,
	// but its close enough for us.
	state, err := rc.GetTreeState()
	if err != nil {
		// If this bailed... we're going to guess its probably misconfiguration
		// not a temporary outage. Lets just bail hard.
		return nil, fmt.Errorf("failed to get initial tree state. Is rekor server configured correctly? Failured caused by %w", err)
	}
	rc.currentIndex = uint(*state.TreeSize)

	return &rc, nil
}

func (rc *impl) GetEntry(index uint) (models.LogEntry, error) {
	entry, err := rc.rekorClient.Entries.GetLogEntryByIndex(&entries.GetLogEntryByIndexParams{LogIndex: int64(index)})
	if err != nil {
		return nil, err
	}
	return entry.GetPayload(), nil
}

func (rc *impl) GetNextEntry() (models.LogEntry, error) {
	entry, err := rc.GetEntry(rc.currentIndex)
	if err != nil {
		return nil, err
	}
	rc.currentIndex++
	return entry, nil
}

func (rc *impl) GetTreeState() (*models.LogInfo, error) {
	glo, err := rc.rekorClient.Tlog.GetLogInfo(&tlog.GetLogInfoParams{})
	if err != nil {
		return nil, err
	}
	return glo.GetPayload(), nil
}

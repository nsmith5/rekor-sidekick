package stdout

import (
	"encoding/json"
	"os"

	"github.com/nsmith5/rekor-sidekick/outputs"
)

const driverName = `stdout`

type impl struct {
	enc *json.Encoder
}

func (i *impl) Send(e outputs.Event) error {
	return i.enc.Encode(e)
}

func (i *impl) Name() string {
	return driverName
}

func New(map[string]interface{}) (outputs.Output, error) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	return &impl{enc}, nil
}

func init() {
	outputs.RegisterDriver(driverName, outputs.CreatorFunc(New))
}

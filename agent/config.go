package agent

import "github.com/nsmith5/rekor-sidekick/policy"

// Config is the data required to configure an agent
type Config struct {
	Server   string                            `yaml:"server"`
	Index    int                               `yaml:"index" default:"-1"`
	Policies []policy.Policy                   `yaml:"policies"`
	Outputs  map[string]map[string]interface{} `yaml:"outputs"`
	Logging  struct {
		Level string
	}
}

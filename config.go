package main

import "github.com/nsmith5/rekor-sidekick/policy"

type config struct {
	RekorServerURL string                            `yaml:"rekorServerURL"`
	Policies       []policy.Policy                   `yaml:"policies"`
	Outputs        map[string]map[string]interface{} `yaml:"outputs"`
}

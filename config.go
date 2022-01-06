package main

type config struct {
	RekorServerURL string                            `yaml:"rekorServerURL"`
	Policies       []policy                          `yaml:"policies"`
	Outputs        map[string]map[string]interface{} `yaml:"outputs"`
}

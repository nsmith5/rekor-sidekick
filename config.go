package main

type config struct {
	RekorServerURL string   `yaml:"rekorServerURL"`
	Policies       []policy `yaml:"policies"`
}

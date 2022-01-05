package main

import (
	"testing"

	"github.com/nsmith5/rekor-sidekick/rekor"
)

func TestPolicy(t *testing.T) {
	tests := map[policy]struct {
		Allow []rekor.LogEntry
		Deny  []rekor.LogEntry
	}{
		policy{
			Name:        `allow-all`,
			Description: ``,
			Body:        "package auth\ndefault allow = true",
		}: {
			Allow: []rekor.LogEntry{
				rekor.LogEntry{
					"spec": map[string]interface{}{
						"foo": "bar",
					},
				},
				rekor.LogEntry{},
			},
			Deny: []rekor.LogEntry{},
		},
		policy{
			Name:        `only x509 signature`,
			Description: ``,
			Body: `package auth
			default auth = false
			allow {
			   format := input.spec.signature.format
			   format == "x509"
			}`,
		}: {
			Allow: []rekor.LogEntry{
				rekor.LogEntry{
					"spec": map[string]interface{}{
						"signature": map[string]interface{}{
							"format": "x509",
						},
					},
				},
			},
			Deny: []rekor.LogEntry{
				rekor.LogEntry{
					"spec": map[string]interface{}{
						"signature": map[string]interface{}{
							"format": "minisign",
						},
					},
				},
			},
		},
	}

	for p, data := range tests {
		t.Run(p.Name, func(t *testing.T) {
			for _, entry := range data.Allow {
				violation, err := p.allowed(entry)
				if err != nil {
					t.Errorf("policy %s failed to check allowed entry with error %s", p.Name, err)
					continue
				}
				if !violation {
					t.Errorf("policy %s denied entry which was expected to be allowed", p.Name)
				}

			}

			for _, entry := range data.Deny {
				violation, err := p.allowed(entry)
				if err != nil {
					t.Errorf("policy %s failed to check allowed entry with error %s", p.Name, err)
					continue
				}
				if violation {
					t.Errorf("policy %s allowed entry which was expected to be denied", p.Name)
				}

			}

		})
	}
}

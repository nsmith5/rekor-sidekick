package policy

import (
	"testing"
)

func TestPolicy(t *testing.T) {
	tests := map[Policy]struct {
		Allow []map[string]interface{}
		Deny  []map[string]interface{}
	}{
		Policy{
			Name:        `allow-all`,
			Description: ``,
			Body:        "package sidekick\ndefault alert = true",
		}: {
			Allow: []map[string]interface{}{
				{
					"spec": map[string]interface{}{
						"foo": "bar",
					},
				},
				{},
			},
			Deny: []map[string]interface{}{},
		},
		Policy{
			Name:        `only x509 signature`,
			Description: ``,
			Body: `package sidekick
			default alert = false
			alert {
			   format := input.spec.signature.format
			   format == "x509"
			}`,
		}: {
			Allow: []map[string]interface{}{
				{
					"spec": map[string]interface{}{
						"signature": map[string]interface{}{
							"format": "x509",
						},
					},
				},
			},
			Deny: []map[string]interface{}{
				{
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
				alert, err := p.Alert(entry)
				if err != nil {
					t.Errorf("policy %s failed to check allowed entry with error %s", p.Name, err)
					continue
				}
				if !alert {
					t.Errorf("policy %s denied entry which was expected to be allowed", p.Name)
				}

			}

			for _, entry := range data.Deny {
				alert, err := p.Alert(entry)
				if err != nil {
					t.Errorf("policy %s failed to check allowed entry with error %s", p.Name, err)
					continue
				}
				if alert {
					t.Errorf("policy %s allowed entry which was expected to be denied", p.Name)
				}

			}

		})
	}
}

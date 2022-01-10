# 🔍 Rekor Sidekick

Rekor Sidekick monitors a Rekor signature transparency log and forwards events
of interest where ever you like.

```
                               ┌─────────────────┐
                               │  Event Policies │
                               └──────┬───▲──────┘
                                      │   │
                             Decision │   │ Should forward entry?
                                      │   │
                                      │   │
                                      │   │                           Outputs
┌─────────────┐              ┌────────▼───┴───────┐
│             │              │                    │                ┌────────────┐
│  Rekor Log  ├──────────────►   Rekor Sidekick   │ ───────────────► Pager Duty │
│             │              │                    │                └────────────┘
└─────────────┘ Pull entries └───────────────┬─┬─┬┘
                                             │ │ │                 ┌────────────┐
                                             │ │ └─────────────────► Stdout     │
                                             │ │                   └────────────┘
                                             │ │
                                             │ │                   ┌────────────┐
                                             │ └───────────────────► Loki       │
                                             │                     └────────────┘
                                             │
                                             │                     ┌────────────┐
                                             └─────────────────────► ...        │
                                                                   └────────────┘
```

## Installation

To install `rekor-sidekick` grab the latest release from our [Github releases
page](https://github.com/nsmith5/rekor-sidekick/releases).

### Verifying a release

Releases are signed and can be verified as follows

```bash
export VERSION="0.1.0"
export ARCH="linux_amd64"
curl -sL "https://github.com/nsmith5/rekor-sidekick/releases/download/v${VERSION}/rekor-sidekick_${VERSION}_${ARCH}.tar.gz" > rekor-sidekick_${VERSION}_${ARCH}.tar.gz
curl -sL "https://github.com/nsmith5/rekor-sidekick/releases/download/v${VERSION}/checksums.txt" > checksums.txt
curl -sL "https://github.com/nsmith5/rekor-sidekick/releases/download/v${VERSION}/checksums.txt.sig" > checksums.txt.sig

export COSIGN_EXPERIMENTAL=1
cosign verify-blob --signature $(cat checksums.txt.sig) checksums.txt
```

The cosign verification step must output sometime to the affect of

```
Certificate is trusted by Fulcio Root CA
Email: []
URI: https://github.com/nsmith5/rekor-sidekick/.github/workflows/release.yml@refs/tags/v0.1.0
Issuer:  https://token.actions.githubusercontent.com
Verified OK
tlog entry verified with uuid: "e530fe7cb3da2ab69535208e54d0c8c63accba35dd75b405c50f23a5093ca712" index: 1029416
```

> NB: The URI should having a version tag matchine `VERSION` and the issuer
> should be https://token.actions.githubusercontent.com.  the tlog entry uuid
> and index are not important.

Finally, hash the release and make sure it matches what you see in `checksums.txt`

```
# Authorized checksums
cat checksums.txt

# Received checksum. Should be in the list of checksums above.
sha256sum rekor-sidekick_${VERSION}_${ARCH}.tar.gz
```

## Usage

`rekor-sidekick` requires a rekor server, alert policies and alert outputs to
be configured. A basic example configuration looks like the following:

```yaml
# config.yaml
server: https://rekor.sigstore.dev
logging:
  level: error
policies:
- name: alert-on-my-email
  description: |
    Alert when an x509 cert with subject email:me@example.com is used
    so sign an entry
  body: |
    package sidekick
    
    import future.keywords.in    
 
    default alert = false

    alert {
      encodedCert := input.spec.signature.publicKey.content
      certs := crypto.x509.parse_certificates(encodedCert)
      emailAddresses := certs[0]["EmailAddresses"]
      some "me@example.com" in emailAddresses
    }
    
outputs:
  stdout:
    enabled: true
```

Launch `rekor-sidekick` by pointing to the config file

```
rekor-sidekick --config /path/to/config.yaml
```

## Configuration

Rekor Sidekick uses a single configuration file with three important sections:

- `server` to point to the Rekor server you want to monitor,
- `policies` to specify which entries you want to alert on, and,
- `outputs` to specify where you want to send your alerts

The `etc` directory contains sample configurations.

## Environment variables

Configuration can also be set using environment variables. They map 1:1 to
configuration fields in the configuration file so that e.g
`.outputs.stdout.enabled` cooresponds to the
`REKOR_SIDEKICK_OUTPUTS_STDOUT_ENABLED` environment variable.

### Writing Alert Policies

Policies are written using the
[Rego](https://www.openpolicyagent.org/docs/latest/policy-language/) policy
language. Some things to remember when writing your policies for Rekor
Sidekick:

- The package name on the policy _must_ be `sidekick`
- Rekor sidekick evalutes the variable `alert` so set it to true in your policy
  if you want to alert on an event
- The base64 decoded contents of the `.[].body` field in a rekor log entry are
  what Rekor sidekick evaluates as input

The best approach to debugging / evalutationg policy is to grab an example log
entry

```
export UUID=<< your example uuid here >>
curl -X GET -H "Accept: application/json" https://rekor.sigstore.dev/api/v1/logs/entries/${UUID} | jq .[].body | base64 -d
```

Paste that data into the [Rego playground](https://play.openpolicyagent.org/)
and iterate on your policy until it behaves how you want.

> NB: you can use `print(x)` to evaluate some data and print to the browser console

### Outputs

**stdout**

The `stdout` driver prints alerts to the console in JSON format. To enable add
the following to your config

```diff
outputs:
+ stdout:
+   enabled: true
```

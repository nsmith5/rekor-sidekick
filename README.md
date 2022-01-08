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

`rekor-sidekick` requires a rekor server, alert policies and alert outputs to be configured. A basic
example configuration looks like the following:

```yaml
# config.yaml
server: https://rekor.sigstore.dev
policies:
- name: alert-all 
  description: |
    Alert all policies alerts on every entry in the transparency log
  body: |
    package sidekick
    default alert = true
outputs:
  stdout:
    enabled: true
```

Launch `rekor-sidekick` by pointing to the config file

```
rekor-sidekick --config /path/to/config.yaml
```

## Configuration

> TODO: write a thorough configuration guide including policy writing and
> description of all the output drivers?

### Writing Alert Policies

### Outputs

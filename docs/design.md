# Rekor Sidekick design

The basic design of Rekor sidekick is to continually pull entries from
a Rekor server, check if the entries are of interest using a configured set
of policies and to forward entries of interest a configured set of outputs

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

Each configured policy should have metadata that helps make the resulting alert
be as actionable as possible. Perhaps a name that is programmatically readable
along with description?

## Event Policy Evaluation

Instead of creating a policy engine, we can an established policy evaluation
engine: Rego policies and Open Policy Agent! This keeps the complexity of
implementation low and we can lean on existing documentation for Rego policies
along with Rekor sidekick specific examples to help folks learn about writing
policies. This approaches keeps policies flexible as well and keeps our
coupling to the Rekor log formats fairly low and shifts that burden on to
policies writers.

## Data structures

We need to define a policy and policy violation:

```go
type Policy struct {
    // short machine readable name
    Name string

    // Background on this policy. Meant for humans. Should help make a 
    // Policy Violation actionable
    Description string

    // Where is this policy defined? Could file:// for on disk or maybe
    // https:// for remote?
    PolicyURI string
}

type PolicyViolation struct {
    // What policy was violated? 
    Policy Policy
    
    // URI of the rekor entry that violates the policy 
    EntryURI string
}
```

## Interfaces

We should make a very simply interface for each of the output implementations so that
contributors can very easily add their own. The following makes sense right now?

```go
type Output interface {
    Send(PolicyViolation) error
}
```

# ADR-001 — Mirror Host Repo Paths Inside The VM

## Status
Approved

## Date
2026-03-17

## Context

The harness needs a path model for repos mounted into the Colima VM.

Options:

- Mirror the host repo paths inside the VM, e.g. `/path/to/workspace/...`
- Mount repos under a VM-specific root such as `/workspace/...`

A VM-specific root makes it clearer whether the engineer is on the host or in
the VM.

Mirrored host paths keep paths consistent between host and VM. Error messages,
logs, and tool output use the same paths in both places. Existing scripts and
workflows are less likely to break.

Docker Sandbox makes the same choice. It exposes the workspace inside the
sandbox at the same absolute path as on the host.

## Decision

Mount repos into the VM at the same absolute paths they use on the host.

Do not introduce a separate `/workspace` root for shared repos.

Use mirrored host paths as the default path model for the harness shell and repo
workflows.

## Consequences

### Benefits

- Host and VM paths stay consistent in logs, errors, and tool output
- Existing scripts and workflows are less likely to break
- The harness does not need path translation logic
- The choice aligns with Docker Sandbox for the same class of problem

### Costs

- It is harder to tell by path alone whether a shell is on the host or in the VM
- The VM boundary is less obvious to the user
- Extra prompt or shell cues may still be useful to make the VM context explicit

## Rejected alternatives

### Mount repos under `/workspace`
This makes the VM boundary clearer. It also introduces path translation and
diverges from a compatibility-oriented choice used by Docker Sandbox for good
reasons.

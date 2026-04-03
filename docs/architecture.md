# Architecture

## Goals

- Keep the host-side CLI small and maintainable
- Keep the runtime model VM-first
- Keep cross-repo access explicit through the harness
- Keep the host-side config model boring

## Host-Side Shape

`harbour` is a single Go binary.

It owns:

- Command dispatch
- Config load and save
- Interactive prompting
- Mount calculation
- Colima start, stop, and ssh orchestration

The host-side config is one JSON file at the Harbour config path returned by
`os.UserConfigDir()`.

## Harness Split

Recommended host-side split:

- `harbour`
  The shareable CLI repo
- `harbour-harness`
  Private local state such as `AGENTS.md`, `repos.yaml`, and `skills/`

Harbour treats `harbour-harness/repos.yaml` as the source of truth for mounted
repo paths.

## VM Runtime

Harbour keeps the VM setup model intentionally simple:

- Work repos are mounted from `repos.yaml`
- Missing repo paths are warned and skipped
- The selected root instruction file is linked at the workspace root
- Custom skills are symlinked into the active agent directory
- The active agent runs directly inside the VM shell

The host-side CLI is Go. The in-VM provision step is still an embedded Bash
script because that boundary is already shell-shaped and mostly imperative
machine setup.

## Path Model

Mounted repos keep the same absolute paths inside the VM as on the host.

This avoids translation logic and keeps logs, diagnostics, and tool output
consistent across host and VM. See [ADR-001](adr/001-mirror-host-repo-paths-inside-the-vm.md).

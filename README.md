# harbour

Colima-backed harness for running Codex across your mounted repos with host isolation.

`harbour` runs the agent inside an isolated VM while keeping it effective across multiple repos.

Your private local configuration lives in `harbour-context`.

## Getting started

1. Set up your context repo from `harbour-context-skeleton`

   See [agent-harbour/harbour-context-skeleton/README.md](https://github.com/agent-harbour/harbour-context-skeleton)

2. Provision the VM

   ```sh
   make provision
   ```

   `make provision` will prompt for `HARBOUR_CONTEXT_HOST_PATH` if needed and
   save it to `~/.config/agent-harbour/env`.

   It will:

   - Start the Colima profile
   - Mount `harbour-context`
   - Mount the work repos from `harbour-context/repos.yaml`
   - Install or update `codex`, `gh`, `make`, and `rg` in the VM
   - Link `AGENTS.md` at `WORKSPACE_ROOT`
   - Sync custom skills from `harbour-context/skills`

3. Start the agent

   ```sh
   make agent
   ```

## Layout

- `Makefile`: Stable entry points
- `config/colima.env`: Colima defaults
- `scripts/`: Provisioning and launch scripts
- `docs/architecture.md`: Runtime model
- `docs/adr/`: Design decisions

## Usage

```sh
make help
make provision
make shell
make agent
make yolo
```

## Notes

The intended runtime is a Colima VM. Codex runs directly in the VM, alongside
repo containers, rather than inside its own container.

Mount only the repo paths declared in `harbour-context/repos.yaml`. Anything
mounted into the VM is intentionally shared with the agent and repo containers.

Each entry in `repos.yaml` is a `host_path` and is mounted read-write.

# harbour

- A colima VM with Codex running in it
- Your repos are mounted into it, defined in `repos.yaml`
- An `AGENTS.md` is mounted
- Skills are mounted
- `make agent`/ `make yolo` puts you onto the agent

## Getting started

1. Set up your context repo

   Your personal context is stored in a separate repo, `harbour-context`.

   See [agent-harbour/harbour-context-skeleton/README.md](https://github.com/agent-harbour/harbour-context-skeleton)

2. Provision the VM

   ```sh
   make provision
   ```

   It will:

   - Prompt for `HARBOUR_CONTEXT_HOST_PATH` if not set
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

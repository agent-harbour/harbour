#!/bin/zsh

set -euo pipefail

SCRIPT_DIR=${0:A:h}
PROJECT_ROOT=${SCRIPT_DIR:h}
COLIMA_ENV="${PROJECT_ROOT}/config/colima.env"
AGENT_CONTEXT_HOST_PATH="${PROJECT_ROOT:h}/agent-context"

if [[ -f "${COLIMA_ENV}" ]]; then
  source "${COLIMA_ENV}"
fi

RUNTIME_ENV="${AGENT_CONTEXT_HOST_PATH}/runtime.env"

if [[ -f "${RUNTIME_ENV}" ]]; then
  source "${RUNTIME_ENV}"
fi

REPOS_FILE="${AGENT_CONTEXT_HOST_PATH}/repos.yaml"

require_var() {
  local name=$1
  if [[ -z "${(P)name:-}" ]]; then
    printf "%s is not set. Configure it in the harness config.\n" "${name}" >&2
    exit 1
  fi
}

repo_lines() {
  require_var AGENT_CONTEXT_HOST_PATH
  if [[ ! -f "${REPOS_FILE}" ]]; then
    printf "%s is missing. Create it in agent-context.\n" "${REPOS_FILE}" >&2
    exit 1
  fi
  awk '
    $1 == "-" && $2 == "name:" {name=$3}
    $1 == "host_path:" {host=$2}
    $1 == "mode:" {mode=$2; printf "%s|%s|%s\n", name, host, mode}
  ' "${REPOS_FILE}"
}

state_root() {
  require_var AGENT_CONTEXT_HOST_PATH
  printf "%s\n" "${AGENT_CONTEXT_HOST_PATH}"
}

bool_flag() {
  local value=$1
  [[ "${value:l}" == "true" ]]
}

colima_status() {
  require_var COLIMA_PROFILE
  colima status -p "${COLIMA_PROFILE}" >/dev/null 2>&1
}

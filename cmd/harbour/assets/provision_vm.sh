set -euo pipefail

selected_agent=$1
run_installer=$2
harbour_harness_agents_path=$3
harbour_harness_skills_dir=$4
harbour_harness_agents_b64=$5
host_uid=$6
host_gid=$7

agent_bin_dir="${HOME}/.local/bin"
codex_path="${agent_bin_dir}/codex"
claude_path="${agent_bin_dir}/claude"
codex_agents_path="${HOME}/.codex/AGENTS.md"
claude_agents_path="${HOME}/.claude/CLAUDE.md"
codex_skills_dir="${HOME}/.codex/skills"
claude_skills_dir="${HOME}/.claude/skills"
case "${selected_agent}" in
  codex)
    other_path=${claude_path}
    agents_path=${codex_agents_path}
    skills_dir=${codex_skills_dir}
    ;;
  claude)
    other_path=${codex_path}
    agents_path=${claude_agents_path}
    skills_dir=${claude_skills_dir}
    ;;
  *) echo "Unsupported agent: ${selected_agent}" >&2; exit 1 ;;
esac

tmpdir=$(mktemp -d)
trap 'rm -rf "${tmpdir}"' EXIT

mkdir -p "${agent_bin_dir}"

sync_skills() {
  local target_skills_dir=$1
  mkdir -p "$(dirname "${target_skills_dir}")"
  rm -rf "${target_skills_dir}"
  if [[ -d "${harbour_harness_skills_dir}" ]]; then
    ln -s "${harbour_harness_skills_dir}" "${target_skills_dir}"
  else
    mkdir -p "${target_skills_dir}"
  fi
}

if ! command -v make >/dev/null 2>&1 || ! command -v rg >/dev/null 2>&1 || ! command -v gh >/dev/null 2>&1 || ! command -v file >/dev/null 2>&1; then
  sudo apt-get update
  sudo apt-get install -y file gh make ripgrep
fi

if [[ ! -f "${harbour_harness_agents_path}" ]]; then
  tmp_agents=$(mktemp)
  trap 'rm -rf "${tmpdir}" "${tmp_agents}"' EXIT
  printf '%s' "${harbour_harness_agents_b64}" | base64 -d > "${tmp_agents}"
  sudo install -o "${host_uid}" -g "${host_gid}" -m 0644 "${tmp_agents}" "${harbour_harness_agents_path}"
fi

if [[ "${run_installer}" == "true" ]]; then
  echo "Running the ${selected_agent} installer for latest..."
  case "${selected_agent}" in
    codex)
      curl -fsSL https://chatgpt.com/codex/install.sh -o "${tmpdir}/installer.sh"
      CODEX_NON_INTERACTIVE=1 CODEX_INSTALL_DIR="${agent_bin_dir}" \
        sh "${tmpdir}/installer.sh" --release latest
      ;;
    claude)
      curl -fsSL https://claude.ai/install.sh -o "${tmpdir}/installer.sh"
      bash "${tmpdir}/installer.sh" latest
      ;;
  esac
fi

export PATH="${agent_bin_dir}:${PATH}"
if ! "${selected_agent}" --version; then
  echo "${selected_agent} --version failed. Run harbour provision and choose to run the installer." >&2
  exit 1
fi

rm -f "${other_path}"
mkdir -p "$(dirname "${agents_path}")"
ln -sfn "${harbour_harness_agents_path}" "${agents_path}"
sync_skills "${skills_dir}"

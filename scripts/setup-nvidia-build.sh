#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_dir="$(cd -- "${script_dir}/.." && pwd)"

config_dir="${XDG_CONFIG_HOME:-${HOME}/.config}/ai-best-model-route"
env_file="${config_dir}/providers.env"
config_file="${config_dir}/config.json"
catalog_file="${config_dir}/nvidia-models.json"
template="${repo_dir}/config/nvidia-build.json"

nvidia_models_url="https://integrate.api.nvidia.com/v1/models"
install=true

usage() {
	cat <<'EOF'
usage: setup-nvidia-build.sh [--configure-only]

Sets up NVIDIA Build / NVIDIA API Catalog for AI Best Model Route.

The script:
  1. accepts NVIDIA_API_KEY from the environment, an existing providers.env,
     or a hidden terminal prompt;
  2. verifies the key against NVIDIA's live /v1/models endpoint;
  3. stores the key only in ~/.config/ai-best-model-route/providers.env (0600);
  4. installs the ready-to-use NVIDIA Build router config;
  5. optionally starts the local Docker router and configures Codex.

Options:
  --configure-only   Validate and write configuration, but do not start Docker
  -h, --help         Show this help
EOF
}

while (($#)); do
	case "$1" in
	--configure-only)
		install=false
		shift
		;;
	-h | --help)
		usage
		exit 0
		;;
	*)
		printf 'error: unknown option: %s\n' "$1" >&2
		usage >&2
		exit 2
		;;
	esac
done

require_command() {
	command -v "$1" >/dev/null 2>&1 || {
		printf 'error: required command not found: %s\n' "$1" >&2
		exit 1
	}
}

for command in curl jq awk grep mktemp chmod mkdir cp mv; do
	require_command "$command"
done

if [[ ! -f "$template" ]]; then
	printf 'error: NVIDIA template not found: %s\n' "$template" >&2
	exit 1
fi

mkdir -p "$config_dir"
chmod 0700 "$config_dir"

key="${NVIDIA_API_KEY:-}"

if [[ -z "$key" && -f "$env_file" ]]; then
	key="$(awk -F= '/^NVIDIA_API_KEY=/{sub(/^NVIDIA_API_KEY=/, ""); print; exit}' "$env_file")"
fi

if [[ -z "$key" ]]; then
	if [[ -t 0 ]]; then
		printf 'Paste your NVIDIA Build API key (input is hidden): '
		IFS= read -r -s key
		printf '\n'
	else
		printf 'error: NVIDIA_API_KEY is not set and no interactive terminal is available.\n' >&2
		printf 'Set NVIDIA_API_KEY or add NVIDIA_API_KEY=... to %s, then rerun.\n' "$env_file" >&2
		exit 1
	fi
fi

if [[ -z "$key" || "$key" == *$'\n'* || "$key" == *$'\r'* ]]; then
	printf 'error: invalid NVIDIA API key value\n' >&2
	exit 1
fi

catalog_tmp="$(mktemp "${config_dir}/.nvidia-models.XXXXXX")"
env_tmp="$(mktemp "${config_dir}/.providers.env.XXXXXX")"
trap 'rm -f -- "${catalog_tmp:-}" "${env_tmp:-}"' EXIT

printf 'Verifying NVIDIA Build API key and discovering account-visible models...\n'

http_code="$(
	curl --silent --show-error --output "$catalog_tmp" --write-out '%{http_code}' --header "Authorization: Bearer $key" --header 'Accept: application/json' "$nvidia_models_url"
)"

if [[ "$http_code" != "200" ]]; then
	printf 'error: NVIDIA model discovery returned HTTP %s\n' "$http_code" >&2
	jq -r '.detail // .message // .error.message // empty' "$catalog_tmp" 2>/dev/null >&2 || true
	exit 1
fi

if ! jq -e '(.data | type == "array") and ([.data[]? | .id? | select(type == "string" and length > 0)] | length > 0)' "$catalog_tmp" >/dev/null; then
	printf 'error: NVIDIA returned no usable model IDs for this key\n' >&2
	exit 1
fi

model_count="$(jq '[.data[]? | .id? | select(type == "string" and length > 0)] | unique | length' "$catalog_tmp")"

# Preserve any existing provider credentials while replacing only NVIDIA_API_KEY.
if [[ -f "$env_file" ]]; then
	grep -v '^NVIDIA_API_KEY=' "$env_file" >"$env_tmp" || true
fi
printf 'NVIDIA_API_KEY=%s\n' "$key" >>"$env_tmp"
chmod 0600 "$env_tmp"
mv -f -- "$env_tmp" "$env_file"

cp -- "$catalog_tmp" "$catalog_file"
chmod 0600 "$catalog_file"

# The NVIDIA template is deterministic; back up a previously generated router
# config before installing it. Existing multi-provider users should normally
# ask Codex to merge NVIDIA into their current configuration instead.
if [[ -f "$config_file" ]]; then
	backup="${config_file}.bak.$(date -u +%Y%m%dT%H%M%SZ)"
	cp -p -- "$config_file" "$backup"
	printf 'Existing router config backed up to %s\n' "$backup"
fi
cp -- "$template" "$config_file"
chmod 0600 "$config_file"

unset NVIDIA_API_KEY
key=""

printf 'NVIDIA Build verified: %s account-visible model(s).\n' "$model_count"
printf 'Credential file: %s\n' "$env_file"
printf 'Router config:   %s\n' "$config_file"
printf 'Catalog snapshot: %s\n' "$catalog_file"

if [[ "$install" != true ]]; then
	printf 'Configuration complete. Docker/Codex installation was skipped.\n'
	exit 0
fi

for command in docker openssl; do
	require_command "$command"
done

replace_args=()
if docker container inspect ai-best-model-route-core >/dev/null 2>&1 ||
	docker container inspect ai-best-model-route-gateway >/dev/null 2>&1; then
	replace_args=(--replace)
fi

printf 'Starting AI Best Model Route and installing the Codex provider...\n'
"${repo_dir}/scripts/setup-local.sh" --config "$config_file" --env-file "$env_file" --model auto --reasoning-effort medium --accept-plaintext-key --accept-new-threads-only "${replace_args[@]}"

printf '\nNVIDIA Build integration is ready.\n'
printf 'Fully quit and reopen Codex, then create a NEW task.\n'
printf 'Use auto for OYYO routing or directly select any NVIDIA Build model shown by the catalog.\n'
printf 'The full account-visible NVIDIA catalog snapshot is in: %s\n' "$catalog_file"

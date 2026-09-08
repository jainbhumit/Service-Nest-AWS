#!/usr/bin/env bash
# Starts DynamoDB Local, initializes the table, and runs the local API server.
# OneDrive/Windows may save CRLF; re-exec with LF so bash options parse correctly.
if [[ -z "${_SN_LF_FIXED:-}" ]] && grep -q $'\r' "$0" 2>/dev/null; then
  _tmp="$(mktemp)"
  tr -d '\r' < "$0" > "${_tmp}"
  export _SN_LF_FIXED=1
  export _SN_SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
  exec bash "${_tmp}" "$@"
fi
set -euo pipefail

SCRIPT_DIR="${_SN_SCRIPT_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)}"
ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
cd "${ROOT}"

if [[ ! -f .env ]]; then
  echo "No .env file found. Copy .env.example to .env and adjust values."
  if [[ -f .env.example ]]; then
    cp .env.example .env
    echo "Created .env from .env.example - update JWT_SECRET before prod use."
  fi
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "docker not found. Install Docker in WSL or start Docker Desktop with WSL integration." >&2
  exit 1
fi

# Export DynamoDB settings from .env for local-init.sh
# Strip CR so Windows-edited .env files still work under bash/WSL.
if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source <(sed 's/\r$//' .env)
  set +a
fi
export DYNAMODB_ENDPOINT="${DYNAMODB_ENDPOINT:-http://localhost:8001}"
export DYNAMODB_TABLE="${DYNAMODB_TABLE:-servicenest}"
export AWS_REGION="${AWS_REGION:-us-east-1}"

echo "Starting Docker Compose ..."
docker compose up -d

echo "Initializing DynamoDB Local table ..."
bash "${ROOT}/scripts/local-init.sh"

echo "Starting local API server ..."
cd "${ROOT}/service-nest"
go run ./cmd/local

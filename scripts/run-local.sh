#!/usr/bin/env bash
# Starts DynamoDB Local, initializes the table, and runs the local API server.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
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

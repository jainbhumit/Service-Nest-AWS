#!/usr/bin/env bash
set -euo pipefail

PROFILE="${AWS_PROFILE:-}"
REGION="${AWS_REGION:-us-east-1}"
STACK_NAME="${STACK_NAME:-serviceNest}"
API_URL="${API_URL:-}"

http_status() {
  local url="$1"
  local method="${2:-GET}"
  local body="${3:-}"
  if [[ -n "${body}" ]]; then
    curl -s -o /dev/null -w "%{http_code}" -X "${method}" -H "Content-Type: application/json" -d "${body}" "${url}"
  else
    curl -s -o /dev/null -w "%{http_code}" -X "${method}" "${url}"
  fi
}

http_body() {
  local url="$1"
  local method="${2:-GET}"
  local body="${3:-}"
  if [[ -n "${body}" ]]; then
    curl -s -X "${method}" -H "Content-Type: application/json" -d "${body}" "${url}"
  else
    curl -s -X "${method}" "${url}"
  fi
}

if [[ -z "${API_URL}" ]]; then
  AWS_ARGS=(cloudformation describe-stacks --stack-name "${STACK_NAME}" --region "${REGION}" \
    --query "Stacks[0].Outputs[?OutputKey=='ServiceNestApiUrl'].OutputValue" --output text --no-cli-pager)
  if [[ -n "${PROFILE}" ]]; then
    AWS_ARGS=(--profile "${PROFILE}" "${AWS_ARGS[@]}")
  fi
  API_URL="$(aws "${AWS_ARGS[@]}")"
fi

API_URL="${API_URL%/}"
if [[ -z "${API_URL}" ]]; then
  echo "Could not resolve ServiceNestApiUrl from stack '${STACK_NAME}'." >&2
  exit 1
fi

echo "Smoke testing ${API_URL}"

health_status="$(http_status "${API_URL}/health")"
if [[ "${health_status}" != "200" ]]; then
  echo "GET /health failed with status ${health_status}" >&2
  exit 1
fi
echo "GET /health -> ${health_status}"

ready_status="$(http_status "${API_URL}/health/ready")"
if [[ "${ready_status}" != "200" ]]; then
  echo "GET /health/ready failed with status ${ready_status}" >&2
  exit 1
fi
echo "GET /health/ready -> ${ready_status}"

login_status="$(http_status "${API_URL}/login" POST "{}")"
if [[ "${login_status}" -lt 400 || "${login_status}" -ge 500 ]]; then
  echo "POST /login expected 4xx, got ${login_status}" >&2
  exit 1
fi

login_body="$(http_body "${API_URL}/login" POST "{}")"
if [[ "${login_body}" != *'"Fail"'* ]]; then
  echo "POST /login response missing standard error envelope" >&2
  exit 1
fi
echo "POST /login -> ${login_status} (error envelope ok)"

echo "Smoke bundle passed for ${API_URL}"

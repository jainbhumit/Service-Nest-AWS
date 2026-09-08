#!/usr/bin/env bash
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

PROFILE="${AWS_PROFILE:-}"
REGION="${AWS_REGION:-us-east-1}"
STACK_NAME="${STACK_NAME:-serviceNest}"
GUIDED=0
NO_CONFIRM=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --profile)
      PROFILE="$2"
      shift 2
      ;;
    --guided)
      GUIDED=1
      shift
      ;;
    --no-confirm)
      NO_CONFIRM=1
      shift
      ;;
    *)
      echo "Unknown argument: $1" >&2
      exit 1
      ;;
  esac
done

if [[ -f "${ROOT}/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source <(sed 's/\r$//' "${ROOT}/.env")
  set +a
fi

JWT_SECRET="${JWT_SECRET:-${SECRET:-}}"
SMTP_PASSWORD="${SMTP_APP_PASSWORD:-${APP_PASSWORD:-}}"
SMTP_FROM="${SMTP_FROM:-}"

if [[ -z "${JWT_SECRET}" ]]; then
  echo "JWT_SECRET (or SECRET) must be set." >&2
  exit 1
fi
if [[ -z "${SMTP_PASSWORD}" ]]; then
  echo "SMTP_APP_PASSWORD (or APP_PASSWORD) must be set." >&2
  exit 1
fi
if [[ -z "${SMTP_FROM}" ]]; then
  echo "SMTP_FROM must be set." >&2
  exit 1
fi

echo "WARNING: This deploy updates PROD stack '${STACK_NAME}' in ${REGION}."
sam build

DEPLOY_ARGS=(
  deploy
  --stack-name "${STACK_NAME}"
  --region "${REGION}"
  --capabilities CAPABILITY_IAM
  --resolve-s3
  --parameter-overrides
  "JwtSecret=${JWT_SECRET}"
  "SmtpAppPassword=${SMTP_PASSWORD}"
  "SmtpFrom=${SMTP_FROM}"
)

if [[ -n "${PROFILE}" ]]; then
  DEPLOY_ARGS+=(--profile "${PROFILE}")
fi
if [[ "${GUIDED}" -eq 1 ]]; then
  DEPLOY_ARGS+=(--guided)
elif [[ "${NO_CONFIRM}" -eq 1 || "${CONFIRM_CHANGESET:-}" == "false" ]]; then
  DEPLOY_ARGS+=(--no-confirm-changeset)
fi

sam "${DEPLOY_ARGS[@]}"

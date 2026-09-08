#!/usr/bin/env bash
set -euo pipefail

ENDPOINT="${DYNAMODB_ENDPOINT:-http://localhost:8001}"
TABLE_NAME="${DYNAMODB_TABLE:-servicenest}"
REGION="${AWS_REGION:-us-east-1}"

echo "Waiting for DynamoDB Local at ${ENDPOINT} ..."
ready=0
for _ in $(seq 1 30); do
  if aws dynamodb list-tables --endpoint-url "${ENDPOINT}" --region "${REGION}" --no-cli-pager >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 1
done

if [ "${ready}" -ne 1 ]; then
  echo "DynamoDB Local is not reachable at ${ENDPOINT}. Run 'docker compose up -d' first." >&2
  exit 1
fi

if aws dynamodb describe-table \
  --endpoint-url "${ENDPOINT}" \
  --region "${REGION}" \
  --table-name "${TABLE_NAME}" \
  --no-cli-pager >/dev/null 2>&1; then
  echo "Table '${TABLE_NAME}' already exists."
  exit 0
fi

echo "Creating table '${TABLE_NAME}' ..."
aws dynamodb create-table \
  --endpoint-url "${ENDPOINT}" \
  --region "${REGION}" \
  --table-name "${TABLE_NAME}" \
  --attribute-definitions AttributeName=PK,AttributeType=S AttributeName=SK,AttributeType=S \
  --key-schema AttributeName=PK,KeyType=HASH AttributeName=SK,KeyType=RANGE \
  --billing-mode PAY_PER_REQUEST \
  --no-cli-pager

echo "Table '${TABLE_NAME}' created."

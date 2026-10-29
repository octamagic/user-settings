#!/usr/bin/env bash

ENDPOINT_URL=${ENDPOINT_URL:-http://dynamodb-local:9000}
REGION=${REGION:-eu-central-1}
USER_SETTINGS_TABLE_NAME=${USER_SETTINGS_TABLE_NAME:-user-settings-user-settings}

# Wait for DynamoDB Local to be available for setup.
until [ "$(curl -s -w '%{http_code}' -o /dev/null "$ENDPOINT_URL")" -eq 400 ]
do
  echo "Waiting for DynamoDB Local..."
  sleep 5
done

aws dynamodb create-table --endpoint-url "$ENDPOINT_URL" --region "$REGION" \
  --table-name "$USER_SETTINGS_TABLE_NAME" \
  --attribute-definitions \
      AttributeName=pk,AttributeType=S \
      AttributeName=sk,AttributeType=S \
  --key-schema \
      AttributeName=pk,KeyType=HASH \
      AttributeName=sk,KeyType=RANGE \
  --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
  --table-class STANDARD

exec /usr/local/bin/aws "$@"

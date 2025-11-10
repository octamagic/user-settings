#!/usr/bin/env bash

ENDPOINT_URL=http://localstack:4566
REGION=eu-central-1

AWS_ACCESS_KEY_ID=local \
  AWS_SECRET_ACCESS_KEY=local \
  aws sqs create-queue --endpoint-url $ENDPOINT_URL --region $REGION --queue-name $WORKER_QUEUE_NAME

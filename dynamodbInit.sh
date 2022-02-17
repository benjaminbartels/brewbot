#!/bin/bash

##  Create tables ##
aws dynamodb create-table \
    --table-name brews-local \
    --attribute-definitions AttributeName=id,AttributeType=S AttributeName=userId,AttributeType=S \
    --key-schema AttributeName=id,KeyType=HASH \
    --global-secondary-indexes \
        "[
            {
                \"IndexName\": \"byUserId\",
                \"KeySchema\": [
                    {\"AttributeName\":\"userId\",\"KeyType\":\"HASH\"}
                ],
                \"Projection\": {
                    \"ProjectionType\":\"ALL\"
                },
                \"ProvisionedThroughput\": {
                    \"ReadCapacityUnits\": 5,
                    \"WriteCapacityUnits\": 5
                }
            }
        ]" \
    --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --endpoint-url http://localhost:8000
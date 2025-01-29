#!/bin/bash

set -e

REGION=${REGION:-"us-west-2"}
FILE=dynamodb_local_latest.tar.gz

if [ ! -f dynamodb/$FILE ]; then
    curl https://s3.${REGION}.amazonaws.com/dynamodb-local/$FILE -q -o dynamodb/$FILE
fi
(cd dynamodb && sha256sum -c dynamodb.sha256)
(cd dynamodb && tar -xvf $FILE)
echo "Done"
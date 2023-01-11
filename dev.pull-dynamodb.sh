#!/bin/bash

REGION=${REGION:-"us-west-2"}
FILE=dynamodb_local_latest.tar.gz

if [ ! -f dynamodb/$FILE ]; then
    curl https://s3.${REGION}.amazonaws.com/dynamodb-local/$FILE -q -o dynamodb/$FILE
fi
rval=$(cd dynamodb && sha256sum -c dynamodb.sha256 && echo $?)
if [ ! "$rval" ]; then
    echo "The target has changed. Pull the latest, commit the hash, and try again."
    exit 1
fi
(cd dynamodb && tar -xvf $FILE)
echo "Done"
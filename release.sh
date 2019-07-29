#!/usr/bin/env bash

name=tfbridge-providers
TAG=`git rev-parse HEAD`

sam package --template-file ${name}.yaml --s3-bucket ${BUCKET} --output-template-file ${name}-packaged.yaml
aws s3 cp ${name}-packaged.yaml s3://${BUCKET}/${TAG}/${name}.yaml

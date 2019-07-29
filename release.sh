#!/usr/bin/env bash

name=tfbridge-providers
TAG=`git rev-parse HEAD`

sam package --template-file ${name}.yaml --s3-bucket ${BUCKET} --s3-prefix releases/${TAG}/artefacts --output-template-file ${name}-packaged.yaml
aws s3 cp ${name}-packaged.yaml s3://${BUCKET}/releases/${TAG}/templates/${name}.yaml

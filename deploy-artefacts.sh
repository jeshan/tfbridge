#!/usr/bin/env bash
set -e

TAG=`cat .version`
TAG=${TAG:-local}
export SAM_CLI_TELEMETRY=0

echo "Got tag ${TAG}"

aws configure set profile.${CLI_PROFILE}.region ${AWS_REGION}
aws configure set profile.${CLI_PROFILE}.credential_source EcsContainer

for filename in dist/*.zip; do
    name=`basename ${filename} .zip`
    sam package --template-file tfbridge/providers/cfn-templates/${name}.yaml --s3-bucket ${BUCKET} --s3-prefix releases/${TAG}/artefacts --output-template-file ${name}-packaged.yaml
    aws s3 cp ${name}-packaged.yaml s3://${BUCKET}/releases/${TAG}/templates/${name}.yaml
done

sceptre --no-colour launch -y main

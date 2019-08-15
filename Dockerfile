FROM golang:1.12-stretch AS tf-builder
ARG GITHUB_TOKEN
ARG BUCKET
ARG AWS_CONTAINER_CREDENTIALS_RELATIVE_URI
ARG AWS_EXECUTION_ENV
ARG AWS_DEFAULT_REGION
ARG AWS_REGION
ARG CLI_PROFILE

ENV GITHUB_TOKEN=${GITHUB_TOKEN} BUCKET=${BUCKET} \
  AWS_CONTAINER_CREDENTIALS_RELATIVE_URI=${AWS_CONTAINER_CREDENTIALS_RELATIVE_URI} \
  AWS_EXECUTION_ENV=${AWS_EXECUTION_ENV} AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION} \
  AWS_REGION=${AWS_REGION} CLI_PROFILE=${CLI_PROFILE}

SHELL ["/bin/bash", "-c"]
WORKDIR /app

RUN apt-get update && apt-get install -y python-dev zip && rm -rf /var/cache/apt
RUN curl https://bootstrap.pypa.io/get-pip.py -o get-pip.py && python get-pip.py
RUN pip install awscli aws-sam-cli sceptre==2.1.5

RUN go mod init github.com/jeshan/tfbridge

COPY tfbridge/lambda tfbridge/lambda
COPY tfbridge/crud tfbridge/crud
#COPY tfbridge/real-tests tfbridge/real-tests
COPY tfbridge/utils tfbridge/utils

RUN go build -o dist/main tfbridge/lambda/main.go

COPY tfbridge/release tfbridge/release
RUN go build -o dist/create-release tfbridge/release/main/create-release.go
RUN go build -o dist/write-build-info tfbridge/release/main/main.go
COPY *.gohtml ./

COPY tfbridge/providers tfbridge/providers

RUN dist/write-build-info
RUN time ./download-dependencies.sh

COPY build-plugins.sh ./
RUN time ./build-plugins.sh

COPY deploy-artefacts.sh ./
COPY templates templates
COPY config config

ENV LC_ALL=C.UTF-8 LANG=C.UTF-8
RUN pip install click==7.0
RUN AWS_REGION=${AWS_REGION} AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION} BUCKET=${BUCKET} CLI_PROFILE=${CLI_PROFILE} ./deploy-artefacts.sh
RUN dist/create-release
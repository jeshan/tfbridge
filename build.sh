#!/usr/bin/env bash

set -e

mkdir -p dist/
rm -rf dist/*

echo "Building main package"
go build -o dist/main tfbridge/lambda/main.go

for name in http github digitalocean gitlab netlify azurerm aws ; do
    echo "Building plugin for ${name}"
    go build -buildmode=plugin -o dist/${name}.so tfbridge/providers/${name}.go
    echo "Packaging plugin for ${name}"
    cd dist/
    zip -9 package-${name}.zip main ${name}.so
    cd ..
    rm dist/${name}.so
done

rm dist/main

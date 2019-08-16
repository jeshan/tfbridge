#!/usr/bin/env bash

for filename in tfbridge/providers/*.go; do
    name=`basename ${filename} .go`
    $(cat download-dependencies.txt | grep terraform-providers-${name})
    echo "Building plugin for ${name}"
    time go build -buildmode=plugin -o dist/${name}.so tfbridge/providers/${name}.go
    if [[ $? -ne 0 ]]; then
        echo "Build plugin for ${name} failed; skipping"
        continue
    fi
    echo "Packaging plugin for ${name}"
    cd dist/
    zip ${name}.zip main ${name}.so
    cd ..
    rm dist/${name}.so
done

rm dist/main

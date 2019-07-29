#!/usr/bin/env bash

provider=${1}
if [[ -z "${provider}" ]]; then
    echo "Specify provider as sole parameter. Value is like terraform-provider-xxx in the https://github.com/terraform-providers org"
    exit 1
fi
provider_file=tfbridge/providers/${provider}.go

cat > ${provider_file} <<EOL
package main

import (
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-${provider}/${provider}"
)

//noinspection ALL
func CreateProvider() (terraform.ResourceProvider, error) {
	provider := ${provider}.Provider()
	rawConfig, err := config.NewRawConfig(map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	conf := terraform.NewResourceConfig(rawConfig)
	err = provider.Configure(conf)
	return provider, err
}
EOL

touch tfbridge/real-tests/${provider}-tests.go

cat << EOF
Configuration file ${provider_file} created. Now,
1. edit it to add custom provider settings.
2. confirm dependency version added in go.mod is the latest release found at https://github.com/terraform-providers/terraform-provider-${provider}/releases
3. add provider to list of binaries built in build.sh
4. run build.sh to download provider confirm builds correctly
5. Write test for ${provider} under at tfbridge/real-tests/${provider}-tests.go
6. Add example provider configuration as custom resource in tfbridge-providers.yaml similar to those already there.
7. Add to list of supported providers in README.md
EOF

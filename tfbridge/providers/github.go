package main

import (
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-github/github"
)

//noinspection GoDuplicate
func CreateProvider() (terraform.ResourceProvider, error) {
	provider := github.Provider()
	rawConfig, err := config.NewRawConfig(map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	conf := terraform.NewResourceConfig(rawConfig)
	err = provider.Configure(conf)
	return provider, err
}

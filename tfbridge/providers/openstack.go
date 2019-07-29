package main

import (
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-openstack/openstack"
)

//noinspection ALL
func CreateProvider() (terraform.ResourceProvider, error) {
	provider := openstack.Provider()
	rawConfig, _ := config.NewRawConfig(map[string]interface{}{})
	conf := terraform.NewResourceConfig(rawConfig)
	err := provider.Configure(conf)
	return provider, err
}

package main

import (
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws"
)

//noinspection GoDuplicate
func CreateProvider() (terraform.ResourceProvider, error) {
	provider := aws.Provider()
	rawConfig, err := config.NewRawConfig(map[string]interface{}{
		"skip_credentials_validation": true,
		"skip_get_ec2_platforms":      true,
		"skip_region_validation":      true,
		"skip_metadata_api_check":     true,
		// "skip_requesting_account_id":  true, may be needed by TF
		"region": "us-east-1",
	})
	if err != nil {
		return nil, err
	}
	conf := terraform.NewResourceConfig(rawConfig)
	err = provider.Configure(conf)
	return provider, err
}

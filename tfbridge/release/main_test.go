package release

import (
	"reflect"
	"strings"
	"testing"
)

func TestListSupportedProviders(t *testing.T) {
	t.Run("paginate-list-org-repos", func(t *testing.T) {
		wantCount := 101
		got := ListSupportedProviders()
		if len(got) <= wantCount {
			t.Errorf("ListSupportedProviders() = %v, expecting at least 101 %v", len(got), wantCount)
		}
		if strings.Index(got[0], "-provider-") >= 0 {
			t.Errorf("ListSupportedProviders() = %v, expecting only provider name, got repo name instead", got[0])
		}
	})
}

func Test_createReleaseNotes(t *testing.T) {
	type args struct {
		tfBridgeVersion  string
		terraformVersion string
		bucket           string
		providers        []ProviderInfo
	}
	bucket := "some-bucket"
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "first",
			args: args{tfBridgeVersion: "v0.1", terraformVersion: "v0.12.x", bucket: bucket, providers: []ProviderInfo{{
				Name:    "aws",
				Version: "latest",
			}}},
			want: `
# tfbridge release v0.1
This release is based on Terraform v0.12.x.


TfBridge v0.1 has the following providers:

## aws (version: latest)
Create an instance by launching a stack:
<a href="https://console.aws.amazon.com/cloudformation/home?#/stacks/new?&templateURL=https://s3.amazonaws.com/some-bucket/releases/v0.1/templates/aws.yaml&stackName=tfbridge-aws" target="_blank"><img src="https://s3.amazonaws.com/cloudformation-examples/cloudformation-launch-stack.png"></a>
*or* run this command:

aws cloudformation create-stack --capabilities CAPABILITY_IAM --stack-name tfbridge-aws --template-url https://s3.amazonaws.com/some-bucket/releases/v0.1/templates/aws.yaml


Upstream provider release notes: [latest](https://github.com/terraform-providers/terraform-provider-aws/releases/latest)

`,
		},
		{
			name: "second",
			args: args{tfBridgeVersion: "v0.2", terraformVersion: "v0.24.5", bucket: bucket, providers: []ProviderInfo{{
				Name:    "aws",
				Version: "latest",
			}, {
				Name:    "xyz",
				Version: "v2.3.4",
			}}},
			want: `
# tfbridge release v0.2
This release is based on Terraform v0.24.5.


TfBridge v0.2 has the following providers:

## aws (version: latest)
Create an instance by launching a stack:
<a href="https://console.aws.amazon.com/cloudformation/home?#/stacks/new?&templateURL=https://s3.amazonaws.com/some-bucket/releases/v0.2/templates/aws.yaml&stackName=tfbridge-aws" target="_blank"><img src="https://s3.amazonaws.com/cloudformation-examples/cloudformation-launch-stack.png"></a>
*or* run this command:

aws cloudformation create-stack --capabilities CAPABILITY_IAM --stack-name tfbridge-aws --template-url https://s3.amazonaws.com/some-bucket/releases/v0.2/templates/aws.yaml


Upstream provider release notes: [latest](https://github.com/terraform-providers/terraform-provider-aws/releases/latest)

## xyz (version: v2.3.4)
Create an instance by launching a stack:
<a href="https://console.aws.amazon.com/cloudformation/home?#/stacks/new?&templateURL=https://s3.amazonaws.com/some-bucket/releases/v0.2/templates/xyz.yaml&stackName=tfbridge-xyz" target="_blank"><img src="https://s3.amazonaws.com/cloudformation-examples/cloudformation-launch-stack.png"></a>
*or* run this command:

aws cloudformation create-stack --capabilities CAPABILITY_IAM --stack-name tfbridge-xyz --template-url https://s3.amazonaws.com/some-bucket/releases/v0.2/templates/xyz.yaml


Upstream provider release notes: [v2.3.4](https://github.com/terraform-providers/terraform-provider-xyz/releases/v2.3.4)

`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createReleaseNotes(tt.args.tfBridgeVersion, tt.args.terraformVersion, tt.args.bucket, tt.args.providers); got != tt.want {
				t.Errorf("createReleaseNotes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readProviderInfo(t *testing.T) {
	tests := []struct {
		name string
		want []ProviderInfo
	}{
		{
			name: "sample-download-dependencies-one.sh",
			want: []ProviderInfo{{
				Name:    "aws",
				Version: "v2.23.0",
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readProviderInfo(tt.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readProviderInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateRelease(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "hey",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			CreateRelease()
		})
	}
}

func Test_getTerraformVersion(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "hey",
			want: "v0.12.5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTerraformVersion(); got != tt.want {
				t.Errorf("getTerraformVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

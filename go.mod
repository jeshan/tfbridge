module github.com/jeshan/tfbridge

go 1.12

require (
	github.com/aws/aws-lambda-go v1.12.0
	github.com/go-critic/go-critic v0.0.0-20181204210945-ee9bf5809ead // indirect
	github.com/google/go-cmp v0.3.0
	github.com/hashicorp/terraform v0.12.5 // local dev only, latest stable will be used in release
	github.com/terraform-providers/terraform-provider-aws v0.0.0-20190726152834-7777619cfbdb
	github.com/terraform-providers/terraform-provider-http v1.1.1
	github.com/terraform-providers/terraform-provider-kubernetes v1.8.1
	google.golang.org/api v0.6.0 // indirect
	k8s.io/apimachinery v0.0.0-20190210215030-4521e64aecd3
)

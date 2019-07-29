package utils

import (
	"testing"
)

func TestGetProviderResourceType(t *testing.T) {
	type args struct {
		cfnType string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "s3-bucket",
			args: args{cfnType: "Custom::TfBridge-resource-aws_s3_bucket"},
			want: "aws_s3_bucket",
		},
		{
			name: "s3-bucket",
			args: args{cfnType: "Custom::TfBridge-resource-AwsS3Bucket"},
			want: "aws_s3_bucket",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetProviderResourceType(tt.args.cfnType); got != tt.want {
				t.Errorf("GetProviderResourceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetProviderName(t *testing.T) {
	type args struct {
		resourceType string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "aws",
			args: args{resourceType: "aws_s3_bucket"},
			want: "aws",
		},
		{
			name: "github",
			args: args{resourceType: "github_repository"},
			want: "github",
		},
		{
			name: "data-aws",
			args: args{resourceType: "aws_iam_user"},
			want: "aws",
		},
		{
			name: "http",
			args: args{resourceType: "http"},
			want: "http",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetProviderName(tt.args.resourceType); got != tt.want {
				t.Errorf("GetProviderResourceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetProviderDataType(t *testing.T) {
	type args struct {
		cfnType string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "iam-user",
			args: args{cfnType: "Custom::TfBridge-data-aws_iam_user"},
			want: "aws_iam_user",
		},
		{
			name: "iam-user",
			args: args{cfnType: "Custom::TfBridge-data-AwsIamUser"},
			want: "aws_iam_user",
		},
		{
			name: "http",
			args: args{cfnType: "Custom::TfBridge-data-http"},
			want: "http",
		},
		{
			name: "http",
			args: args{cfnType: "Custom::TfBridge-resource-whatever"},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetProviderDataType(tt.args.cfnType); got != tt.want {
				t.Errorf("GetProviderResourceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareAttributes(t *testing.T) {
	type args struct {
		importedAttributes map[string]string
		properties         map[string]interface{}
		logicalResourceID  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "user-with-tag-list",
			args: args{
				importedAttributes: map[string]string{
					"arn":       "arn:aws:iam::406883123076:user/s3_public",
					"id":        "s3_public",
					"name":      "s3_public",
					"path":      "/",
					"tags.%":    "1",
					"tags.a":    "b",
					"unique_id": "AIDAV5PA7X6CGFP5BMIK7",
				},
				properties: map[string]interface{}{
					"arn":  "arn:aws:iam::406883123076:user/s3_public",
					"id":   "s3_public",
					"name": "s3_public",
					"path": "/",
					// "tags.%":    "1", // TODO: remove this line
					"tags":      map[string]interface{}{"a": "b"},
					"unique_id": "AIDAV5PA7X6CGFP5BMIK7",
				},
				logicalResourceID: "User",
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CompareAttributes(tt.args.importedAttributes, tt.args.properties, tt.args.logicalResourceID); (err != nil) != tt.wantErr {
				t.Errorf("CompareAttributes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

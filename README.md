# tfbridge

![Badge](https://codebuild.us-east-1.amazonaws.com/badges?uuid=eyJlbmNyeXB0ZWREYXRhIjoiRlNPU3hDRHFrT2oxZXpzeGVJd0xDRDBldE1JNWxkbTkzWFRNY0NSUWY2dFRzaC8wR3NmeWVoSERHQlVWL1djWS9ibUgyVmVGNXZrbEIvRm1OYkgzWldnPSIsIml2UGFyYW1ldGVyU3BlYyI6Im9mc2N4STBmaEF6MjBiRDQiLCJtYXRlcmlhbFNldFNlcmlhbCI6MX0%3D&branch=master)

Create, import and manage (virtually) **any** Terraform resource in (virtually) any provider on AWS CloudFormation (via custom resources).

## Why would somebody want this?
- Cloudformation devotees: Stop feeling jealous that Terraform has more features.
- Configure your postgres RDS instance with database (privileges, databases, schema, etc)
- Use *tfbridge* to bring existing resources under CloudFormation management.
- Leverage your existing CloudFormation skills to deploy to multiple providers: combine Github, Netlify and AWS in one template.
- Skip waiting for the CloudFormation to provide native support for new services and features.
- Terraform users: No longer need to mess with TF state files; state is handled by Cloudformation.

# What it is
*tfbridge* is a bunch of Cloudformation custom resources backed by serverless functions. It is fairly provider agnostic so that it's easier to support as many of them as possible.

## Features
- [x] Multi-provider. The following are currently available and it's trivial to add more. The ticked ones are tested by me:
  - [x] AWS
  - [ ] Azure
  - [x] DigitalOcean
  - [x] Github
  - [x] Gitlab
  - [ ] Google Cloud Platform
  - [x] Http
  - [ ] Kubernetes
  - [x] Netlify
  - [ ] OpenStack
  - [ ] PostgreSQL
- [x] Terraform data sources
- [x] Import resources (just like in Terraform)
  - [x] Strict mode: *tfbridge* can check that you declared all properties correctly.
- [ ] variable interpolation e.g `${var.self.whatever}`.

## Usage
1. Deploy the stack using the template on the releases page. It shows how to create the serverless functions that can provision resources in the supported providers. Use the parameters to pass in your credentials to the various providers, e.g your digital ocean access token. Note the function names of the deployed resources. You will use it in the next step.
2. Next, create custom resources in the following format. The next section has some examples:
  - resource:
  ```yaml
    MyResource:
      Type: Custom::TfBridge-resource-$RESOURCE
      Properties:
        ServiceToken: !Sub arn:aws:lambda:${AWS::Region}:${AWS::AccountId}:function:$STACK_NAME-$PROVIDER
      param1: val1 # as documented in the resource's Terraform docs.
      param2: val2
  ```
  - data source: 
  ```yaml
  MyData:
    Type: Custom::TfBridge-data-$DATA_SOURCE
    Properties:
      ServiceToken: !Sub arn:aws:lambda:${AWS::Region}:${AWS::AccountId}:function:$STACK_NAME-$DATA_SOURCE
      param1: val1 # as documented in the data source's Terraform docs.
      param2: val2
  ``` 
3. Optional: To deploy to same providers using different credentials, relaunch a new stack using the same template in the first step. Supply it with new credentials or edit the template to customise it further.

### Attributes
Terraform's resources returns several attributes, e.g A `github_repository` returns full_name, git_clone_url, [etc](https://www.terraform.io/docs/providers/github/r/repository.html#attributes-reference). In Terraform, you would refer to them as `${github_repository.my_repo.git_clone_url`. With *tfbridge*, you do it as such: `!GetAtt MyRepo.git_clone_url`.

## Example resources
You can try the following snippets. They are intended to work as similarly to the original Terraform project as much as possible:

An HTTP data source:
```yaml
  HttpData:
    Type: Custom::TfBridge-data-http
    Properties:
      ServiceToken: !Sub arn:aws:lambda:${AWS::Region}:${AWS::AccountId}:function:tfbridge-http
      # as documented here https://www.terraform.io/docs/providers/http/data_source.html
      url: https://checkpoint-api.hashicorp.com/v1/check/terraform
      request_headers:
        Accept: application/json

```

A Netlify site:
```yaml
  NetlifySite:
    Type: Custom::TfBridge-resource-netlify_site
    Properties:
      ServiceToken: !Sub arn:aws:lambda:${AWS::Region}:${AWS::AccountId}:function:tfbridge-netlify
      # as documented here https://www.terraform.io/docs/providers/netlify/r/netlify_site.html
      #name: some-custom-name
      repo:
        - command: gulp build
          dir: dist/
          provider: github
          repo_branch: master
          repo_path: jeshan/cloudformation-checklist
```

Importing an existing AWS IAM user. Set `TFBRIDGE_MODE` = Import and `TFBRIDGE_ID` to the ID of the resource to be imported:
```yaml
  User:
    Type: Custom::TfBridge-resource-aws_iam_user
    Properties:
      TFBRIDGE_MODE: Import
      TFBRIDGE_ID: some_user_name_to_import
      ServiceToken: !Sub arn:aws:lambda:${AWS::Region}:${AWS::AccountId}:function:tfbridge-aws
```

Importing the same user, but this time strictly checking that all properties have been properly mapped. Set `TFBRIDGE_MODE` to **ImportStrict**:
```yaml
  User:
    Type: Custom::TfBridge-resource-aws_iam_user
    Properties:
      TFBRIDGE_MODE: ImportStrict
      TFBRIDGE_ID: some_user_name_to_import
      ServiceToken: !Sub arn:aws:lambda:${AWS::Region}:${AWS::AccountId}:function:tfbridge-aws
      arn: !Sub arn:aws:iam::${AWS::AccountId}:user/some_user_name_to_import
      id: some_user_name_to_import
      name: some_user_name_to_import
      path: /
      unique_id: AIDAV5PA7X6CGFEXAMPLE
```

When in doubt, check the relevant Terraform docs.

## Configuring the providers
*tfbridge* leverages configuration features already supported by Terraform. Since TF resources are configurable via environment variables, you can configure their respective serverless function with environment variables.
To know the full list of available env vars, check the `provider.go` file for them. e.g `NETLIFY_TOKEN` is the env var to set for Netlify: https://github.com/terraform-providers/terraform-provider-netlify/blob/master/netlify/provider.go#L13-L24

## Notes
- If you don't see your favourite provider, raise an [issue with this link](https://github.com/jeshan/tfbridge/issues/new?title=Add%20support%20for%20provider%20$x&body=Please%20support%20provider%20$x.%20%20It%27s%20available%20at%20the%20following%20link:https://github.com/terraform-providers/terraform-provider-$x)

- Please remember that **this is still experimental software**. Do not use it in production yet.

## Forkers
Code is released under the Simplified BSD Licence. Fork and hack away!

Install Go 1.12 locally. Run `build.sh` to compile code for the various providers. Example resources that you can deploy to test is found on this page or in the `custom-resources.yaml` template.

You can deploy a similar deployment pipeline via the [templates/infrastructure.yaml](templates/infrastructure.yaml) file. The Codebuild project in it contains the exact build steps. In case you can't build successfully, check the exact steps in it. Otherwise, raise an issue.

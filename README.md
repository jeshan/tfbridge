# tfbridge

![Badge](https://codebuild.us-east-1.amazonaws.com/badges?uuid=eyJlbmNyeXB0ZWREYXRhIjoiRlNPU3hDRHFrT2oxZXpzeGVJd0xDRDBldE1JNWxkbTkzWFRNY0NSUWY2dFRzaC8wR3NmeWVoSERHQlVWL1djWS9ibUgyVmVGNXZrbEIvRm1OYkgzWldnPSIsIml2UGFyYW1ldGVyU3BlYyI6Im9mc2N4STBmaEF6MjBiRDQiLCJtYXRlcmlhbFNldFNlcmlhbCI6MX0%3D&branch=master)

Create, import and manage (virtually) **any** Terraform resource in (virtually) any provider on AWS CloudFormation (via custom resources).

## Why on earth would you do this?
- Cloudformation devotees: Stop feeling jealous that Terraform has more features.
- Use *tfbridge* to bring existing resources under CloudFormation management.
- Leverage your existing CloudFormation skills to deploy to multiple providers: combine Github, Netlify and AWS in one template.
- Skip waiting for the CloudFormation to provide native support for new services and features.
- Terraform users: No longer need to mess with TF state files; state is handled by Cloudformation.

## Features
- [x] Multi-provider. The following are currently available on the releases page and it's trivial to add more. The ticked ones are tested by me:
  - [x] AWS
  - [ ] Azure
  - [x] DigitalOcean
  - [x] Github
  - [x] Gitlab
  - [x] Http
  - [x] Netlify
- [x] Terraform data sources
- [x] Import resources (just like in Terraform)
  - [x] Strict mode: *tfbridge* can check that you declared all properties correctly.
- [ ] variable interpolation e.g `${var.self.whatever}`. To refer to other resources, use CFN's `!GetAtt ${Resource.Prop}` syntax.

### Notes
- If you don't see your favourite provider, raise an [issue with this link](https://github.com/jeshan/tfbridge/issues/new?title=Add%20support%20for%20provider%20$x&body=Please%20support%20provider%20$x.%20%20It%27s%20available%20at%20the%20following%20link:https://github.com/terraform-providers/terraform-provider-$x)

- Please remember that **this is still experimental software**. Do not use it in production yet.

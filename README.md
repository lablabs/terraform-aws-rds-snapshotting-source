# terraform-aws-rds-snapshotting-source

[<img src="ll-logo.png">](https://lablabs.io/)

We help companies build, run, deploy and scale software and infrastructure by embracing the right technologies and principles. Check out our website at https://lablabs.io/

---

## Description

Terraform module for creating RDS snapshots and shipping them to another account for auditing/archival purposes. Works with [terraform-aws-rds-snapshotting-target](https://github.com/lablabs/terraform-aws-rds-snapshotting-target) which has to be deployed to the targer (audit) account.

## Features

- creates a lambda function that periodically takes RDS snapshots

## Usage

### Requirements

No requirements.

### Providers

| Name | Version |
|------|---------|
| aws | n/a |

### Modules

No Modules.

### Resources

| Name |
|------|
| [aws_cloudwatch_event_rule](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_event_rule) |
| [aws_cloudwatch_event_target](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_event_target) |
| [aws_cloudwatch_log_group](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/cloudwatch_log_group) |
| [aws_iam_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy) |
| [aws_iam_role](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role) |
| [aws_iam_role_policy_attachment](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_role_policy_attachment) |
| [aws_lambda_function](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_function) |
| [aws_lambda_permission](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_permission) |

### Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| db\_cluster\_name | n/a | `any` | n/a | yes |
| db\_snapshot\_retention\_days | n/a | `any` | n/a | yes |
| lambda\_schedule | n/a | `string` | `"cron(0 * * * ? *)"` | no |
| region | n/a | `any` | n/a | yes |
| target\_account | n/a | `any` | n/a | yes |
| target\_sns\_topic\_arn | n/a | `any` | n/a | yes |

### Outputs

No output.
## Contributing and reporting issues

Feel free to create an issue in this repository if you have questions, suggestions or feature requests.

### Building lambda package

- use `asdf` to install necessary go version
- compile go code
- zip compiled code and push to the repository

```bash
asdf install

GOARCH=amd64 GOOS=linux go build -o bootstrap main.go

zip lambda.zip bootstrap
```

## License

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

See [LICENSE](LICENSE) for full details.

    Licensed to the Apache Software Foundation (ASF) under one
    or more contributor license agreements.  See the NOTICE file
    distributed with this work for additional information
    regarding copyright ownership.  The ASF licenses this file
    to you under the Apache License, Version 2.0 (the
    "License"); you may not use this file except in compliance
    with the License.  You may obtain a copy of the License at

      https://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing,
    software distributed under the License is distributed on an
    "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
    KIND, either express or implied.  See the License for the
    specific language governing permissions and limitations
    under the License.

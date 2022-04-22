# Introduction

Shell Script to import the existing AWS organization from the master account

Terraform import will import organization configuration from an AWS account.

```
root_id=$(aws organizations list-roots --query 'Roots[*].[Id]' --output=text)
root_name=$(aws organizations list-roots --query 'Roots[*].[Name]' --output=text)
terraform import aws_organizations_organization."$root_name" "$root_id"
```

Inorder for this to work properly, a corresponding terraform resource is required in the terraform file. At a minimum the following configuration is required for the import to work.

```
resource "aws_organizations_organization" Root {}
```

UPDATE: The tool has been uplifted to go lang
This provides an opportunity to retrieve the resource configuration for future maintainenance.

## Requirements

Working AWS CLI is required to gather details

## Usage

script to import the state:

```bash
./exec_imports.sh
```

go script to import the state:

```bash
./ go run *.go
```

script to scrub the state file:

```bash
./scrub_state.sh
```

# bcf-terraform
This repository contains the terraform provider implementation for BigSwitch's
Big Cloud Fabric (BCF) Controller product. The provider plugin helps the DevOps
team to build, change and version their BCF Controller configurations easily via
Terraform. Essentially managing BCF via the DevOps friendly Infra-As-Code
model.


# Using the provider

## Installing the provider plugin
To install the provider plugin, copy over the provider plugin to the directory
from where you shall be running the terraform commands. Then, initialize
terraform so it detects the provider
```bash
# ls terraform-provider-bcf*
terraform-provider-bcf_v0.5.0*

# terraform init
<snip>
* provider.bcf: version = "~> 0.5"
Terraform has been successfully initialized!
<snip>
```


## Using the provider plugin
The provider plugin can be used to configure a number of constructs on the BCF
Controller, like eVPCs, Segments, Subnets, Switches, etc.

Sample terraform template can be found in ```"./examples/"```


# Developer Notes
Golang version -
```bash
go version
> go version go1.11.5 darwin/amd64
```

go mod is used to track external library dependencies. To update a dependency,
run the following
```bash
# go get -u <pkg>
# go mod vendor
```

Terraform library version -
```bash
# cat go.mod | grep terraform
	github.com/hashicorp/terraform v0.11.13
```

## Building the provider plugin
To build the provider, simply run the following command
```bash
make docker-terraform
```
On a successful build, the binary (terraform-provider-bcf_version) shall be
in the ```'./bin/'``` directory.

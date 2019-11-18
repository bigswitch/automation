# Downloading Ansible
Download Ansible for your system:
https://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html?extIdCarryOver=true&sc_cid=701f2000001OH7YAAW

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

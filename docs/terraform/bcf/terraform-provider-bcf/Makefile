OUTDIR=bin
MODULENAME=github.com/bigswitch/bcf-terraform
VER=0.5.0

GOFLAGS=-ldflags '-s -w' -mod vendor
GOTESTFLAGS=-coverprofile=c.out $(GO_MOD_VENDOR_FLAG)

DOCKERIMG=bigswitch/bcf-cni-builder:1.12.5

.NOTPARALLEL:

default: terraform

terraform:
		go build $(GOFLAGS) -o $(OUTDIR)/terraform-provider-bcf_v$(VER) $(MODULENAME)/provider

docker-terraform:
		docker run --rm -v `pwd`:/code $(DOCKERIMG) make terraform

test:
		# go test -v $(MODULENAME)/bcfrestclient $(GOTESTFLAGS)
		go test -v $(MODULENAME)/provider $(GOTESTFLAGS)

clean:
		rm -rf ./$(OUTDIR)

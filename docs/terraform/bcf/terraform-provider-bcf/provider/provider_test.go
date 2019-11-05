package main

import (
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"testing"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider


func init() {
	os.Setenv(resource.TestEnvVar, "true")

	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"bcf": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("BCF_IP"); v == "" {
		t.Fatal("BCF_IP must be set for acceptance tests")
	}
	if v := os.Getenv("BCF_ACCESS_TOKEN"); v == "" {
		t.Fatal("BCF_ACCESS_TOKEN must be set for acceptance tests")
	}
}
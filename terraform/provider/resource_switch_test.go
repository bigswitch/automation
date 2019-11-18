package main

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

func TestAccSwitch_complete(t *testing.T) {
	resourceName := "bcf_switch.test"

	swName1 := "leafa"
	fabricRole1 := "leaf"
	mac1 := "a0:04:d3:f3:b1:2b"
	leafGroup1 := "lg1"
	desc := "desc"

	swName2 := "leafa"
	fabricRole2 := "spine"
	mac2 := "a0:04:d3:f3:c1:2b"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			// testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSwitch(swName1, fabricRole1, mac1, leafGroup1, desc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", swName1),
					resource.TestCheckResourceAttr(resourceName, "fabric_role", fabricRole1),
					resource.TestCheckResourceAttr(resourceName, "mac_address", mac1),
					resource.TestCheckResourceAttr(resourceName, "leaf_group", leafGroup1),
					resource.TestCheckResourceAttr(resourceName, "description", desc),
				),
			},
			{
				Config: testAccDataSwitch(swName2, fabricRole1, mac1, leafGroup1, desc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", swName2),
					resource.TestCheckResourceAttr(resourceName, "fabric_role", fabricRole1),
					resource.TestCheckResourceAttr(resourceName, "mac_address", mac1),
					resource.TestCheckResourceAttr(resourceName, "leaf_group", leafGroup1),
					resource.TestCheckResourceAttr(resourceName, "description", desc),
				),
			},
			{
				Config: testAccDataSwitch(swName1, fabricRole2, mac2, "", desc),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", swName1),
					resource.TestCheckResourceAttr(resourceName, "fabric_role", fabricRole2),
					resource.TestCheckResourceAttr(resourceName, "mac_address", mac2),
					resource.TestCheckResourceAttr(resourceName, "description", desc),
				),
			},
		},
	})
}

func testAccDataSwitch(name, role, mac, leafGroup, desc string) string {
	return fmt.Sprintf(`
resource "bcf_switch" "test" {
  name = "%s"
  fabric_role = "%s",
  mac_address = "%s",
  leaf_group = "%s",
  description = "%s"
}
`, name, role, mac, leafGroup, desc)
}

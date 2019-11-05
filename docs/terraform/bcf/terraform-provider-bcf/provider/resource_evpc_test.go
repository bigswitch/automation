package main

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

func TestAccEVpc_complete(t *testing.T) {
	resourceName := "bcf_evpc.test"

	evpcName1 := "terraform1"
	desc1 := "desc1"

	evpcName2 := "terraform2"
	desc2 := "desc2"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			// testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataEVpc(evpcName1, desc1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", evpcName1),
					resource.TestCheckResourceAttr(resourceName, "description", desc1),
				),
			},
			{
				Config: testAccDataEVpc(evpcName1, desc2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", evpcName1),
					resource.TestCheckResourceAttr(resourceName, "description", desc2),
				),
			},
			{
				Config: testAccDataEVpc(evpcName2, desc2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", evpcName2),
					resource.TestCheckResourceAttr(resourceName, "description", desc2),
				),
			},
			{
				Config: testAccDataEVpc(evpcName2, desc1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", evpcName2),
					resource.TestCheckResourceAttr(resourceName, "description", desc1),
				),
			},
		},
	})
}


func testAccDataEVpc(evpc string, desc string) string {
	return fmt.Sprintf(`
resource "bcf_evpc" "test" {
  name = "%s"
  description = "%s"
}
`, evpc, desc)
}

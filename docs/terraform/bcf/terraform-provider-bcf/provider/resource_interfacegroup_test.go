package main

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"strings"
	"testing"
)

func TestAccInterfaceGroup_complete(t *testing.T) {
	resourceName := "bcf_interface_group.test"

	evpcName1 := "terraform1"
	mode1 := "static"
	desc1 := "desc1"
	swIfMapList1 := make([]map[string]string, 0)
	swIfMapList1 = append(swIfMapList1, map[string]string{"switch": "leaf1", "interface": "eth1"})
	swIfMapList1 = append(swIfMapList1, map[string]string{"switch": "leaf2", "interface": "eth1"})

	evpcName2 := "terraform2"
	desc2 := "desc2"
	swIfMapList2 := make([]map[string]string, 0)
	swIfMapList2 = append(swIfMapList2, map[string]string{"switch": "leaf2", "interface": "eth2"})

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			// testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataInterfaceGroup(evpcName1, mode1, swIfMapList1, desc1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", evpcName1),
					resource.TestCheckResourceAttr(resourceName, "mode", mode1),
					resource.TestCheckResourceAttr(resourceName, "description", desc1),
					resource.TestCheckResourceAttr(resourceName, "switch_interface_list.#", "2"),
				),
			},
			{
				Config: testAccDataInterfaceGroup(evpcName2, mode1, swIfMapList1, desc1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", evpcName2),
					resource.TestCheckResourceAttr(resourceName, "mode", mode1),
					resource.TestCheckResourceAttr(resourceName, "description", desc1),
					resource.TestCheckResourceAttr(resourceName, "switch_interface_list.#", "2"),
				),
			},
			{
				Config: testAccDataInterfaceGroup(evpcName2, mode1, swIfMapList2, desc2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", evpcName2),
					resource.TestCheckResourceAttr(resourceName, "mode", mode1),
					resource.TestCheckResourceAttr(resourceName, "description", desc2),
					resource.TestCheckResourceAttr(resourceName, "switch_interface_list.#", "1"),
				),
			},
		},
	})
}


func testAccDataInterfaceGroup(evpc string, mode string, swIfList []map[string]string, desc string) string {
	swIfListBytes,_ := json.Marshal(swIfList)
	swIfListStr := string(swIfListBytes)
	swIfListStr = strings.ReplaceAll(swIfListStr,":", "=")
	s := fmt.Sprintf(`
resource "bcf_interface_group" "test" {
  name = "%s"
  mode = "%s"
  switch_interface_list = %s
  description = "%s"
}
`, evpc, mode, swIfListStr, desc)
return s
}

# BCF Terraform Fabric Examples

## Add/Delete Fabric Switches

```terraform
provider "bcf" {
  credentials_file_path = "~/.bcf/credentials"
}

resource "bcf_switch" "my-leafa" {
  name = "rack1-leafa"
  fabric_role = "leaf",
  mac_address = "a0:04:d3:f3:a1:2b",
  leaf_group = "rack1-leaf-group",
  description = "tor leaf-a in rack1"
}

resource "bcf_switch" "my-leafb" {
  name = "rack1-leafb"
  fabric_role = "leaf",
  mac_address = "a0:04:d3:f3:a1:2c",
  leaf_group = "rack1-leaf-group",
  description = "tor leaf-b in rack1"
}
```
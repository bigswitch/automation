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

resource "bcf_evpc" "my-evpc" {
  name = "terraform"
  description = "evpc terraform"
}

resource "bcf_segment" "my-evpc-segment-1" {
  name = "frontend"
  evpc = "${bcf_evpc.my-evpc.name}"
  description = "segment for frontend apps"
  subnets = [
    "10.10.10.1/24",
    "10.10.11.1/24"]
}

resource "bcf_segment" "my-evpc-segment-2" {
  name = "backend"
  evpc = "${bcf_evpc.my-evpc.name}"
  description = "segment for backend apps"
  subnets = [
    "20.20.20.1/24",
    "20.20.21.1/24"]
}

resource "bcf_interface_group" "my-interface-group" {
  name = "ig-terraform"
  mode = "static"
  switch_interface_list = [
    {
      "switch" = "${bcf_switch.my-leafa.name}",
      "interface" = "ethernet32"
    },
    {
      "switch" = "${bcf_switch.my-leafb.name}",
      "interface" = "ethernet32"
    }
  ]
  description = "interface-group ig-terraform"
}

resource "bcf_memberrule" "my-memberrule-in-segment-1" {
  evpc = "${bcf_segment.my-evpc-segment-1.evpc}"
  segment = "${bcf_segment.my-evpc-segment-1.name}"
  interface_group = "${bcf_interface_group.my-interface-group.name}"
  vlan = 10
}

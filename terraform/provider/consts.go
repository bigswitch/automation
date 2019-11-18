package main

const (
	LogFile = "/var/log/bcfterraform.log"

	ResourceEVPC                 = "bcf_evpc"
	ResourceSegment              = "bcf_segment"
	ResourceSwitch               = "bcf_switch"
	ResourceInterfaceGroup       = "bcf_interface_group"
	ResourceMemberRuleIfaceGroup = "bcf_memberrule"

	AttrIP           = "ip"
	AttrAccessToken  = "access_token"
	AttrCredFilePath = "credentials_file_path"

	AttrName                = "name"
	AttrDesc                = "description"
	AttrSubnets             = "subnets"
	AttrTenant              = "evpc"
	AttrSegment             = "segment"
	AttrInterfaceGroup      = "interface_group"
	AttrVlan                = "vlan"
	AttrMacAddr             = "mac_address"
	AttrFabricRole          = "fabric_role"
	AttrLeafGroup           = "leaf_group"
	AttrShutdown            = "shutdown"
	AttrMode                = "mode"
	AttrSwitchInterfaceList = "switch_interface_list"
	AttrSwitch              = "switch"
	AttrInterface           = "interface"

	AttrValueFabricRoleLeaf  = "leaf"
	AttrValueFabricRoleSpine = "spine"
)
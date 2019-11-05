/*
 * Copyright 2019 Big Switch Networks, Inc.
 */

package bcfrestclient

import (
	"testing"
)

var (
	testBcfIP = "10.5.6.3"
	testUser  = "admin"
	testOrig  = "test"
	testToken = "3rJpmi0jWWbzZRCsqzvl7jRrLX-l4UVa"

	tName   = "TestTenant"
	sName   = "TestSegment"
	sName2  = "TestSegment2"
	cidr    = "220.200.200.1/24"
	cidr2   = "220.200.201.1/24"
	epName  = "TestEP"
	mac     = "FA:AB:AB:AB:AB:AB"
	ip      = "220.200.200.2/24"
	swName  = "TestSwitch"
	swName2 = "TestSwitch2"
	swIface = "iface-name"
	iface   = "TestIface"
)

func Test_Complete(t *testing.T) {
	bcf := New(testBcfIP, DefBcfPort, testUser, testToken, PasswdEncToken, testOrig, PluginTypeTerraform)

	err := bcf.CreateTenant(tName, tName, "Tenant Description")
	if err != nil {
		t.Error("Unexpected failure")
	}
	tInfo, err := bcf.GetTenant(tName)
	if err != nil {
		t.Error("Unexpected failure")
	}
	if tInfo.Name != tName {
		t.Error("Unexpected failure")
	}

	err = bcf.CreateSegment(tName, sName, sName, "Segment Description")
	if err != nil {
		t.Error("Unexpected failure")
	}

	sInfo, err := bcf.GetSegment(tName, sName)
	if err != nil {
		t.Error("Unexpected failure")
	}
	if sInfo.Name != sName {
		t.Error("Unexpected failure")
	}

	err = bcf.CreateSegmentIface(tName, sName)
	if err != nil {
		t.Error("Unexpected failure")
	}

	segIface, err := bcf.GetSegmentIface(tName, sName)
	if err != nil || segIface.Segment != sName || segIface.Private != true {
		t.Error("Unexpected failure")
	}

	err = bcf.CreateSegmentIfaceSubnet(tName, sName, cidr)
	if err != nil {
		t.Error("Unexpected failure")
	}
	segIface1, err := bcf.GetSegmentIfaceSubnet(tName, sName, cidr)
	if err != nil || segIface1.CIDR != cidr {
		t.Error("Unexpected failure")
	}

	err = bcf.CreateSegmentIface(tName, sName2)
	if err != nil {
		t.Error("Unexpected failure")
	}
	err = bcf.CreateSegmentIfaceSubnet(tName, sName2, cidr2)
	if err != nil {
		t.Error("Unexpected failure")
	}
	segIface2, err := bcf.GetSegmentIfaceSubnet(tName, sName2, cidr2)
	if err != nil || segIface2.CIDR != cidr2 {
		t.Error("Unexpected failure")
	}

	bcfCidrs, err := bcf.GetAllSubnetsForSegment(tName, sName)
	if err != nil || len(bcfCidrs) != 1 {
		t.Error("Unexpected failure")
	}

	err = bcf.DeleteSegmentIfaceSubnet(tName, sName, cidr)
	if err != nil {
		t.Error("Unexpected failure")
	}
	err = bcf.DeleteSegmentIfaceSubnet(tName, sName, cidr2)
	if err != nil {
		t.Error("Unexpected failure")
	}

	bcfCidrs, err = bcf.GetAllSubnetsForSegment(tName, sName)
	if err != nil || len(bcfCidrs) != 0 {
		t.Error("Unexpected failure")
	}

	err = bcf.DeleteSegmentIface(tName, sName)
	if err != nil {
		t.Error("Unexpected failure")
	}

	err = bcf.DeleteSegmentIface(tName, sName2)
	if err != nil {
		t.Error("Unexpected failure")
	}

	err = bcf.DeleteSegment(tName, sName)
	if err != nil {
		t.Error("Unexpected failure")
	}

	err = bcf.DeleteTenant(tName)
	if err != nil {
		t.Error("Unexpected failure")
	}
}

func Test_FabricErrors(t *testing.T) {
	bcf := New(testBcfIP, DefBcfPort, testUser, testToken, PasswdEncToken, testOrig, PluginTypeTerraform)
	err := bcf.GetHealth()
	if err != nil {
		t.Error("Unexpected failure")
	}
}

func Test_SwitchConfig(t *testing.T) {
	bcf := New(testBcfIP, DefBcfPort, testUser, testToken, PasswdEncToken, testOrig, PluginTypeTerraform)
	swName := "test-switch"
	mac := "d0:04:d3:f3:a1:3d"
	fabricRole := "leaf"
	leafGroup := "test-leafgroup-a"
	desc := "test switch"
	shutdown := false

	err := bcf.CreateSwitch(swName, mac, fabricRole, leafGroup, desc, shutdown)
	if err != nil {
		t.Error("Unexpected failure")
	}
	swInfo, err := bcf.GetSwitch(swName)
	if err != nil {
		t.Error("Unexpected failure")
	}
	if swInfo.FabricRole != fabricRole || swInfo.MacAddr != mac || swInfo.LeafGroup != leafGroup {
		t.Error("Unexpected failure")
	}

	err = bcf.DeleteSwitch(swName)
	if err != nil {
		t.Error("Unexpected failure")
	}
}

func Test_InterfaceGroup(t *testing.T) {
	bcf := New(testBcfIP, DefBcfPort, testUser, testToken, PasswdEncToken, testOrig, PluginTypeTerraform)
	igName := "ig-terraform"
	igMode := "static"
	desc := "test interface group"
	swIfMap := make(map[string]string, 0)
	swIfMap["leafa"] = "ethernet1"
	swIfMap["leafb"] = "ethernet1"

	err := bcf.CreateInterfaceGroup(igName, igMode, swIfMap, desc)
	if err != nil {
		t.Error("Unexpected failure", err)
	}

	igInfo, err := bcf.GetInterfaceGroup(igName)
	if err != nil {
		t.Error("Unexpected failure", err)
	}
	if igInfo.Name != igName || igInfo.Mode != igMode || igInfo.Description != igInfo.Description || len(igInfo.MemberInterface) != len(swIfMap) {
		t.Error("Unexpected failure")
	}

	err = bcf.DeleteInterfaceGroup(igName)
	if err != nil {
		t.Error("Unexpected failure", err)
	}
}

func Test_MemberRuleInterfaceGroup(t *testing.T) {
	bcf := New(testBcfIP, DefBcfPort, testUser, testToken, PasswdEncToken, testOrig, PluginTypeTerraform)

	tName := "terraform"
	sName := "segment1"
	igName := "ig-terraform"
	vlan := 10

	err := bcf.CreateTenant(tName, tName, "Tenant Description")
	if err != nil {
		t.Error("Unexpected failure")
	}

	err = bcf.CreateSegment(tName, sName, sName, "Segment Description")
	if err != nil {
		t.Error("Unexpected failure")
	}

	err = bcf.CreateMemberRuleIfaceGroup(tName, sName, igName, vlan)
	if err != nil {
		t.Error("Unexpected failure", err)
	}

	info, err := bcf.GetMemberRuleIfaceGroup(tName, sName, igName, vlan)
	if err != nil || info.InterfaceGroup != igName || info.Vlan != vlan {
		t.Error("Unexpected failure")
	}

	err = bcf.DeleteMemberRuleIfaceGroup(tName, sName, igName, vlan)
	if err != nil {
		t.Error("Unexpected failure")
	}

	bcf.DeleteSegment(tName, sName)
	bcf.DeleteTenant(tName)

	return
}
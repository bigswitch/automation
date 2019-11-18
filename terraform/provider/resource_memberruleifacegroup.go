package main

import (
	"fmt"
	"github.com/bigswitch/bcf-terraform/bcfrestclient"
	"github.com/bigswitch/bcf-terraform/logger"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceMemberRuleInterfaceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceMemberRuleInterfaceGroupCreate,
		Read:   resourceMemberRuleInterfaceGroupRead,
		Update: resourceMemberRuleInterfaceGroupUpdate,
		Delete: resourceMemberRuleInterfaceGroupDelete,

		Schema: map[string]*schema.Schema{
			AttrTenant: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			AttrSegment: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			AttrInterfaceGroup: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			AttrVlan: &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func genMemberRuleInterfaceGroupId(tName, sName, igName string, vlan int) string {
	return fmt.Sprintf("%s%s%s%d", tName, sName, igName, vlan)
}

func createMemberRuleInterfaceGroup(client bcfrestclient.BCFRestClient, tName string, sName string, igName string, vlan int) error {
	return client.CreateMemberRuleIfaceGroup(tName, sName, igName, vlan)
}

func resourceMemberRuleInterfaceGroupCreate(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	igName := d.Get(AttrInterfaceGroup).(string)
	tName := d.Get(AttrTenant).(string)
	sName := d.Get(AttrSegment).(string)
	vlan := d.Get(AttrVlan).(int)
	id := genMemberRuleInterfaceGroupId(tName, sName, igName, vlan)

	logger.Debugf("Create called for member-rule interface-group %s\n", igName)

	err := createMemberRuleInterfaceGroup(bcfclient, tName, sName, igName, vlan)

	if err != nil {
		d.SetId("")
		return err
	}
	d.SetId(id)
	return nil
}

func resourceMemberRuleInterfaceGroupRead(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	igName := d.Get(AttrInterfaceGroup).(string)
	tName := d.Get(AttrTenant).(string)
	sName := d.Get(AttrSegment).(string)
	vlan := d.Get(AttrVlan).(int)

	logger.Debugf("Read called for member-rule interface-group %s\n", igName)

	_, err := bcfclient.GetMemberRuleIfaceGroup(tName, sName, igName, vlan)
	if err != nil {
		if bcfrestclient.IsBCFConnectivityErr(err) {
			return err
		}
		d.SetId("")
		return nil
	}
	// In the case of member-rule, either config is persent or not. The internal
	// attributes can't be modified
	return nil
}

func resourceMemberRuleInterfaceGroupUpdate(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	igName := d.Get(AttrInterfaceGroup).(string)
	tName := d.Get(AttrTenant).(string)
	sName := d.Get(AttrSegment).(string)
	vlan := d.Get(AttrVlan).(int)

	logger.Debugf("Update called for member-rule interface-group %s\n", igName)

	id := genMemberRuleInterfaceGroupId(tName, sName, igName, vlan)

	// If there are any changes to the config, delete the old config and create the new one
	oldName := igName
	oldTName := tName
	oldSName := sName
	oldVlan := vlan

	if d.HasChange(AttrInterfaceGroup) {
		old, _ := d.GetChange(AttrInterfaceGroup)
		oldName = old.(string)
	}
	if d.HasChange(AttrTenant) {
		old, _ := d.GetChange(AttrTenant)
		oldTName = old.(string)
	}
	if d.HasChange(AttrSegment) {
		old, _ := d.GetChange(AttrSegment)
		oldSName = old.(string)
	}
	if d.HasChange(AttrVlan) {
		old, _ := d.GetChange(AttrVlan)
		oldVlan = old.(int)
	}

	deleteMemberRuleInterfaceGroup(bcfclient, oldTName, oldSName, oldName, oldVlan)
	err := createMemberRuleInterfaceGroup(bcfclient, tName, sName, igName, vlan)
	if err != nil {
		d.SetId("")
		return err
	}

	d.SetId(id)
	return resourceMemberRuleInterfaceGroupRead(d, m)
}

func deleteMemberRuleInterfaceGroup(client bcfrestclient.BCFRestClient, tName string, sName string, igName string, vlan int) error {
	return client.DeleteMemberRuleIfaceGroup(tName, sName, igName, vlan)
}

func resourceMemberRuleInterfaceGroupDelete(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	igName := d.Get(AttrInterfaceGroup).(string)
	tName := d.Get(AttrTenant).(string)
	sName := d.Get(AttrSegment).(string)
	vlan := d.Get(AttrVlan).(int)

	logger.Debugf("Delete called for member-rule interface-group %s\n", igName)

	deleteMemberRuleInterfaceGroup(bcfclient, tName, sName, igName, vlan)
	d.SetId("")
	return nil
}

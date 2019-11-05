package main

import (
	"github.com/bigswitch/bcf-terraform/bcfrestclient"
	"github.com/bigswitch/bcf-terraform/logger"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceInterfaceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceInterfaceGroupCreate,
		Read:   resourceInterfaceGroupRead,
		Update: resourceInterfaceGroupUpdate,
		Delete: resourceInterfaceGroupDelete,

		Schema: map[string]*schema.Schema{
			AttrName: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			AttrMode: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			AttrSwitchInterfaceList: &schema.Schema{
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
				Required: true,
			},
			AttrDesc: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func genInterfaceGroupId(name string) string {
	return name
}

func createInterfaceGroup(client bcfrestclient.BCFRestClient, name string, mode string, swIfMap map[string]string, desc string) error {
	return client.CreateInterfaceGroup(name, mode, swIfMap, desc)
}

func resourceInterfaceGroupCreate(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	name := d.Get(AttrName).(string)
	mode := d.Get(AttrMode).(string)
	swIfaceMapList := d.Get(AttrSwitchInterfaceList).(*schema.Set)
	desc := d.Get(AttrDesc).(string)

	logger.Debugf("Create called for interface-group %s\n", name)

	id := genInterfaceGroupId(name)

	swIfaceMap := make(map[string]string, 0)
	for _, entry := range swIfaceMapList.List() {
		mapEntry, okMap := entry.(map[string]interface{})
		if !okMap {
			logger.Warnf("Ignoring unexpected config %+v\n", entry)
			continue
		}
		if len(mapEntry) < 1 {
			continue
		}
		swName, okSw := mapEntry[AttrSwitch].(string)
		ifName, okIf := mapEntry[AttrInterface].(string)
		if !okSw || !okIf {
			logger.Warnf("Ignoring unexpected config %+v\n", mapEntry)
			continue
		}
		swIfaceMap[swName] = ifName
	}

	err := createInterfaceGroup(bcfclient, name, mode, swIfaceMap, desc)
	if err != nil {
		d.SetId("")
		return err
	}
	d.SetId(id)
	return nil
}

func resourceInterfaceGroupRead(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	name := d.Get(AttrName).(string)

	logger.Debugf("Read called for interface-group %s\n", name)

	info, err := bcfclient.GetInterfaceGroup(name)
	if err != nil {
		if bcfrestclient.IsBCFConnectivityErr(err) {
			return err
		}
		d.SetId("")
		return nil
	}

	swIfMapList := make([]map[string]string, 0)
	for _, entry := range info.MemberInterface {
		swIfMap := make(map[string]string, 0)
		swIfMap[AttrSwitch] = entry.Switch
		swIfMap[AttrInterface] = entry.Interface
		swIfMapList = append(swIfMapList, swIfMap)
	}

	d.Set(AttrName, info.Name)
	d.Set(AttrMode, info.Mode)
	d.Set(AttrDesc, info.Description)
	d.Set(AttrSwitchInterfaceList, swIfMapList)

	return nil
}

func resourceInterfaceGroupUpdate(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	name := d.Get(AttrName).(string)

	logger.Debugf("Update called for interface-group %s\n", name)

	id := genInterfaceGroupId(name)
	var err error
	if d.HasChange(AttrName) {
		// The name of interface-group has changed, delete the old interface-group
		// and create the new one
		oldName, _ := d.GetChange(AttrName)
		deleteInterfaceGroup(bcfclient, oldName.(string))
		err = resourceInterfaceGroupCreate(d, m)
	} else if d.HasChange(AttrMode) || d.HasChange(AttrDesc) || d.HasChange(AttrSwitchInterfaceList) {
		// Mode or Desc or SwitchInterface of interface-group has changed, simply update it
		err = resourceInterfaceGroupCreate(d, m)
	}

	if err != nil {
		d.SetId("")
		return err
	}

	d.SetId(id)
	return resourceInterfaceGroupRead(d, m)
}

func deleteInterfaceGroup(client bcfrestclient.BCFRestClient, name string) error {
	return client.DeleteInterfaceGroup(name)
}

func resourceInterfaceGroupDelete(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	name := d.Get(AttrName).(string)

	logger.Debugf("Delete called for interface-group %s\n", name)

	deleteInterfaceGroup(bcfclient, name)

	d.SetId("")
	return nil
}

package main

import (
	"github.com/bigswitch/bcf-terraform/bcfrestclient"
	"github.com/bigswitch/bcf-terraform/logger"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceEVPC() *schema.Resource {
	return &schema.Resource{
		Create: resourceEVPCCreate,
		Read:   resourceEVPCRead,
		Update: resourceEVPCUpdate,
		Delete: resourceEVPCDelete,

		Schema: map[string]*schema.Schema{
			AttrName: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			AttrDesc: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func createEVPC(client bcfrestclient.BCFRestClient, tName string, id string, desc string) error {
	return client.CreateTenant(tName, id, desc)
}

func resourceEVPCCreate(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	tName := d.Get(AttrName).(string)
	desc := d.Get(AttrDesc).(string)
	id := tName

	logger.Debugf("Create called for eVPC %s\n", tName)

	err := createEVPC(bcfclient, tName, id, desc)
	if err != nil {
		d.SetId("")
		return err
	}
	d.SetId(id)
	return nil
}

func resourceEVPCRead(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	tName := d.Get(AttrName).(string)

	logger.Debugf("Read called for eVPC %s\n", tName)

	tInfo, err := bcfclient.GetTenant(tName)
	if err != nil {
		if bcfrestclient.IsBCFConnectivityErr(err) {
			return err
		}
		d.SetId("")
		return nil
	}

	d.Set(AttrName, tInfo.Name)
	d.Set(AttrDesc, tInfo.Description)
	return nil
}

func resourceEVPCUpdate(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	desc := d.Get(AttrDesc).(string)
	tName := d.Get(AttrName).(string)
	id := tName

	logger.Debugf("Update called for eVPC %s\n", tName)

	var err error
	if d.HasChange(AttrName) {
		// Delete old tenant and create a new tenant with config
		oldValue, _ := d.GetChange(AttrName)
		deleteEVPC(bcfclient, oldValue.(string))

		err = createEVPC(bcfclient, tName, id, desc)
	} else if d.HasChange(AttrDesc) {
		// Update existing tenant with config
		err = createEVPC(bcfclient, tName, id, desc)
	}

	if err != nil {
		d.SetId("")
		return err
	}

	d.SetId(id)
	return resourceEVPCRead(d, m)
}

func deleteEVPC(client bcfrestclient.BCFRestClient, tName string) error {
	return client.DeleteTenant(tName)
}

func resourceEVPCDelete(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	tName := d.Get(AttrName).(string)

	logger.Debugf("Delete called for eVPC %s\n", tName)

	deleteEVPC(bcfclient, tName)

	d.SetId("")
	return nil
}
package main

import (
	"fmt"
	"github.com/bigswitch/bcf-terraform/bcfrestclient"
	"github.com/bigswitch/bcf-terraform/logger"
	"github.com/hashicorp/terraform/helper/schema"
	"regexp"
)

func resourceSwitch() *schema.Resource {
	return &schema.Resource{
		Create: resourceSwitchCreate,
		Read:   resourceSwitchRead,
		Update: resourceSwitchUpdate,
		Delete: resourceSwitchDelete,

		Schema: map[string]*schema.Schema{
			AttrName: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			AttrMacAddr: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if !regexp.MustCompile(`^([0-9a-fA-F]{2}[:-]){5}([0-9a-fA-F]{2})$`).MatchString(value) {
						errors = append(errors, fmt.Errorf("invalid %q format", k))
					}
					return
				},
			},
			AttrFabricRole: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if value != AttrValueFabricRoleLeaf && value != AttrValueFabricRoleSpine {
						errors = append(errors, fmt.Errorf("invalid %q specified for switch", k))
					}
					return
				},
			},
			AttrLeafGroup: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			AttrDesc: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			AttrShutdown: &schema.Schema{
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
		},
	}
}

func genSwitchId(swName string) string {
	return swName
}

func createSwitch(client bcfrestclient.BCFRestClient, swName, mac, fabricRole, leafGroup, description string, shutdown bool) error {
	return client.CreateSwitch(swName, mac, fabricRole, leafGroup, description, shutdown)
}

func resourceSwitchCreate(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	swName := d.Get(AttrName).(string)
	mac := d.Get(AttrMacAddr).(string)
	fabricRole := d.Get(AttrFabricRole).(string)
	leafGroup := d.Get(AttrLeafGroup).(string)
	desc := d.Get(AttrDesc).(string)
	shutdown := d.Get(AttrShutdown).(bool)

	logger.Debugf("Create called for switch %s\n", swName)

	id := genSwitchId(swName)

	err := createSwitch(bcfclient, swName, mac, fabricRole, leafGroup, desc, shutdown)
	if err != nil {
		d.SetId("")
		return err
	}
	d.SetId(id)
	return nil
}

func resourceSwitchRead(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	swName := d.Get(AttrName).(string)

	logger.Debugf("Read called for switch %s\n", swName)

	info, err := bcfclient.GetSwitch(swName)
	if err != nil {
		if bcfrestclient.IsBCFConnectivityErr(err) {
			return err
		}
		d.SetId("")
		return nil
	}
	d.Set(AttrName, info.Name)
	d.Set(AttrMacAddr, info.MacAddr)
	d.Set(AttrFabricRole, info.FabricRole)
	d.Set(AttrLeafGroup,info.LeafGroup)
	d.Set(AttrDesc, info.Description)
	d.Set(AttrShutdown,info.Shutdown)
	return nil
}

func resourceSwitchUpdate(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	swName := d.Get(AttrName).(string)
	mac := d.Get(AttrMacAddr).(string)
	fabricRole := d.Get(AttrFabricRole).(string)
	leafGroup := d.Get(AttrLeafGroup).(string)
	desc := d.Get(AttrDesc).(string)
	shutdown := d.Get(AttrShutdown).(bool)

	logger.Debugf("Update called for switch %s\n", swName)

	id := genSwitchId(swName)

	var err error
	if d.HasChange(AttrName) {
		// The switch name has changed, delete the old config and create new one
		oldSwName, _ := d.GetChange(AttrName)
		deleteSwitch(bcfclient, oldSwName.(string))

		err = createSwitch(bcfclient, swName, mac, fabricRole, leafGroup, desc, shutdown)
	} else if d.HasChange(AttrMacAddr) || d.HasChange(AttrFabricRole) || d.HasChange(AttrLeafGroup) ||
		d.HasChange(AttrDesc) || d.HasChange(AttrShutdown) {
		// The switch config has changed, update its config
		err = createSwitch(bcfclient, swName, mac, fabricRole, leafGroup, desc, shutdown)
	}

	if err != nil {
		d.SetId("")
		return err
	}

	d.SetId(id)
	return resourceSwitchRead(d, m)
}

func deleteSwitch(client bcfrestclient.BCFRestClient, swName string) error {
	return client.DeleteSwitch(swName)
}

func resourceSwitchDelete(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	swName := d.Get(AttrName).(string)

	logger.Debugf("Delete called for switch %s\n", swName)

	deleteSwitch(bcfclient, swName)

	d.SetId("")
	return nil
}
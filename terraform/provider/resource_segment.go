package main

import (
	"github.com/bigswitch/bcf-terraform/bcfrestclient"
	"github.com/bigswitch/bcf-terraform/logger"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceSegment() *schema.Resource {
	return &schema.Resource{
		Create: resourceSegmentCreate,
		Read:   resourceSegmentRead,
		Update: resourceSegmentUpdate,
		Delete: resourceSegmentDelete,

		Schema: map[string]*schema.Schema{
			AttrName: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			AttrTenant: &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			AttrDesc: &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			AttrSubnets: &schema.Schema{
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
		},
	}
}

func genSegmentId(tName, sName string) string {
	return tName + "." + sName
}

func createSegment(client bcfrestclient.BCFRestClient, tName string, sName string, id string, desc string) error {
	return client.CreateSegment(tName, sName, id, desc)
}

func createSegmentIface(client bcfrestclient.BCFRestClient, tName string, sName string) error {
	return client.CreateSegmentIface(tName, sName)
}

func createSegmentIfaceSubnet(client bcfrestclient.BCFRestClient, tName string, sName string, cidr string) error {
	return client.CreateSegmentIfaceSubnet(tName, sName, cidr)
}

func resourceSegmentCreate(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	sName := d.Get(AttrName).(string)
	tName := d.Get(AttrTenant).(string)
	desc := d.Get(AttrDesc).(string)
	subnets := d.Get(AttrSubnets).(*schema.Set)

	logger.Debugf("Create called for segment %s in eVPC %s\n", sName, tName)

	id := genSegmentId(tName, sName)

	d.Partial(true)
	err := createSegment(bcfclient, tName, sName, id, desc)
	if err != nil {
		d.SetId("")
		return err
	}
	d.SetPartial("segment")

	if subnets.Len() > 0 {
		err = createSegmentIface(bcfclient, tName, sName)
		if err != nil {
			d.SetId("")
			return err
		}
	}
	d.SetPartial("segment-lr")

	for _, subnet := range subnets.List() {
		err = createSegmentIfaceSubnet(bcfclient, tName, sName, subnet.(string))
		if err != nil {
			d.SetId("")
			return err
		}
	}
	d.SetPartial("segment-lr-subnets")

	d.Partial(false)
	d.SetId(id)
	return nil
}

func resourceSegmentRead(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	sName := d.Get(AttrName).(string)
	tName := d.Get(AttrTenant).(string)

	logger.Debugf("Read called for segment %s in eVPC %s\n", sName, tName)

	info, err := bcfclient.GetSegment(tName, sName)
	if err != nil {
		if bcfrestclient.IsBCFConnectivityErr(err) {
			return err
		}
		d.SetId("")
		return nil
	}

	bcfCidrs, err := bcfclient.GetAllSubnetsForSegment(tName, sName)
	if err != nil {
		if bcfrestclient.IsBCFConnectivityErr(err) {
			return err
		}
		d.SetId("")
		return nil
	}

	d.Set(AttrName, info.Name)
	d.Set(AttrDesc, info.Description)
	d.Set(AttrSubnets, bcfCidrs)
	return nil
}

func resourceSegmentUpdate(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	sName := d.Get(AttrName).(string)
	tName := d.Get(AttrTenant).(string)
	desc := d.Get(AttrDesc).(string)
	subnets := d.Get(AttrSubnets).(*schema.Set)

	logger.Debugf("Update called for segment %s in eVPC %s\n", sName, tName)

	id := genSegmentId(tName, sName)

	var err error
	if d.HasChange(AttrTenant) {
		// The tenant to which the segment belongs has changed, delete the old
		// segment config and create the new segment
		oldTenant, _ := d.GetChange(AttrTenant)
		deleteSegment(bcfclient, oldTenant.(string), sName)
		deleteSegmentIface(bcfclient, oldTenant.(string), sName)

		err = resourceSegmentCreate(d, m)
	} else if d.HasChange(AttrName) {
		// The segment name has changed, delete the old segment and create the
		// new segment
		oldSegment, _ := d.GetChange(AttrName)
		deleteSegment(bcfclient, tName, oldSegment.(string))
		deleteSegmentIface(bcfclient, tName, oldSegment.(string))

		err = resourceSegmentCreate(d, m)
	} else if d.HasChange(AttrDesc) {
		// Update existing segment with config
		err = createSegment(bcfclient, tName, sName, id, desc)
	} else if d.HasChange(AttrSubnets) {
		if subnets.Len() == 0 {
			err = deleteSegmentIface(bcfclient, tName, sName)
		} else {
			oldSubnets, _ := d.GetChange(AttrSubnets)
			oldSubnetsCast := oldSubnets.(*schema.Set)

			addedSubnets, _, deletedSubnets := getSegmentIfaceSubnetDiff(subnets.List(), oldSubnetsCast.List())

			d.Partial(true)
			d.SetPartial("segment")

			if oldSubnetsCast.Len() == 0 {
				err = createSegmentIface(bcfclient, tName, sName)
				if err != nil {
					d.SetId("")
					return err
				}
			}
			d.SetPartial("segment-lr")

			for _, delSubnet := range deletedSubnets {
				deleteSegmentIfaceSubnet(bcfclient, tName, sName, delSubnet)
			}
			for _, addSubnet := range addedSubnets {
				err = createSegmentIfaceSubnet(bcfclient, tName, sName, addSubnet)
				if err != nil {
					d.SetId("")
					return err
				}
			}
			d.SetPartial("segment-lr-subnets")
			d.Partial(false)
		}
	}

	if err != nil {
		d.SetId("")
		return err
	}

	d.SetId(id)
	return resourceSegmentRead(d, m)
}

func getSegmentIfaceSubnetDiff(newSubnets []interface{}, oldSubnets []interface{}) ([]string, []string, []string) {
	var added, same, deleted []string
	isNewSubnet := make(map[string]bool, 0)

	for _, newSubnet := range newSubnets {
		isNewSubnet[newSubnet.(string)] = true
	}
	for _, oldSubnet := range oldSubnets {
		if _, ok := isNewSubnet[oldSubnet.(string)]; ok {
			isNewSubnet[oldSubnet.(string)] = false
		} else {
			deleted = append(deleted, oldSubnet.(string))
		}
	}
	for subnet, isNew := range isNewSubnet {
		if isNew {
			added = append(added, subnet)
		} else {
			same = append(same, subnet)
		}
	}
	return added, same, deleted
}

func deleteSegment(client bcfrestclient.BCFRestClient, tName string, sName string) error {
	return client.DeleteSegment(tName, sName)
}

func deleteSegmentIface(client bcfrestclient.BCFRestClient, tName string, sName string) error {
	return client.DeleteSegmentIface(tName, sName)
}

func deleteSegmentIfaceSubnet(client bcfrestclient.BCFRestClient, tName string, sName string, cidr string) error {
	return client.DeleteSegmentIfaceSubnet(tName, sName, cidr)
}

func resourceSegmentDelete(d *schema.ResourceData, m interface{}) error {
	bcfclient := m.(bcfrestclient.BCFRestClient)
	sName := d.Get(AttrName).(string)
	tName := d.Get(AttrTenant).(string)

	logger.Debugf("Delete called for segment %s in eVPC %s\n", sName, tName)

	deleteSegment(bcfclient, tName, sName)
	deleteSegmentIface(bcfclient, tName, sName)

	d.SetId("")
	return nil
}
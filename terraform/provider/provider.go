package main

import (
	"errors"
	"github.com/bigswitch/bcf-terraform/bcfrestclient"
	"github.com/bigswitch/bcf-terraform/logger"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

const (
	DefBcfCredsFilePath = "~/.bcf/credentials"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			AttrIP: &schema.Schema{
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("BCF_IP", ""),
				Optional:    true,
			},
			AttrAccessToken: &schema.Schema{
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("BCF_ACCESS_TOKEN", ""),
				Optional:    true,
			},
			AttrCredFilePath: &schema.Schema{
				Type:        schema.TypeString,
				DefaultFunc: schema.EnvDefaultFunc("BCF_CREDS_FILE_PATH", DefBcfCredsFilePath),
				Optional:    true,
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			ResourceEVPC:                 resourceEVPC(),
			ResourceSegment:              resourceSegment(),
			ResourceSwitch:               resourceSwitch(),
			ResourceInterfaceGroup:       resourceInterfaceGroup(),
			ResourceMemberRuleIfaceGroup: resourceMemberRuleInterfaceGroup(),
		},

		ConfigureFunc: configureProvider,
	}
}

func genBcfCredsConfig(d *schema.ResourceData) (*bcfrestclient.BcfCredsConfig, error) {
	cfg := &bcfrestclient.BcfCredsConfig{}

	cfg.Default.Ip = d.Get(AttrIP).(string)
	cfg.Default.AccessToken = d.Get(AttrAccessToken).(string)
	if len(cfg.Default.Ip) < 1 || len(cfg.Default.AccessToken) < 1 {
		// Read the credentials file to get bcf info
		credFilePath := d.Get(AttrCredFilePath).(string)
		logger.Infof("Credential info for BCF Controller not provided in template. Trying to read from file %s\n", credFilePath)

		err := ParseConfigfile(credFilePath, cfg)
		if err != nil {
			return nil, errors.New("failed parsing credentials from file " + credFilePath)
		}
	}
	return cfg, nil
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	cfg, err := genBcfCredsConfig(d)
	if err != nil {
		logger.Fatalf("Credential info for BCF Controller not found: %s\n", err)
		return nil, err
	}

	bcfclient := bcfrestclient.NewFromCredsConfig(cfg)
	// Validate connectivity and credentials
	err = bcfclient.GetHealth()
	if err != nil {
		logger.Fatalf("Connectivity check with BCF Controller failed: %s\n", err)
		return nil, err
	}
	return bcfclient, nil
}

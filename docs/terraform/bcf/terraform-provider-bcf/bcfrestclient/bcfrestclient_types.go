/*
 * Copyright 2019 Big Switch Networks, Inc.
 */

package bcfrestclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

var ErrBCFConnTimedOut = errors.New("connection request to bcf timed-out")
var ErrBCFCtrlFailOver = errors.New("bcf controller failover detected")
var ErrBCFAuth = errors.New("bcf authorization failed")

var EmptyData = bytes.NewBuffer([]byte(``))

// REST method constants
const (
	Orig = "terraform"

	PluginTypeTerraform = "terraform"

	MethodGet    = "GET"
	MethodPost   = "POST"
	MethodPut    = "PUT"
	MethodDelete = "DELETE"
)

const (
	ResApiBase = "/api/v1"
	ResApiData = "/data/controller"

	ResLogin              = "/auth/login"
	ResErrors             = "/applications/bcf/info/errors/fabric"
	ResTenant             = "/applications/bcf/tenant[name=\"%s\"]"
	ResSegment            = "/applications/bcf/tenant[name=\"%s\"]/segment[name=\"%s\"]"
	ResSegmentIface       = "/applications/bcf/tenant[name=\"%s\"]/logical-router/segment-interface[segment=\"%s\"]"
	ResSegmentIfaceSubnet = "/applications/bcf/tenant[name=\"%s\"]/logical-router/segment-interface[segment=\"%s\"]/ip-subnet[ip-cidr=\"%s\"]"
	ResTenantLR           = "/applications/bcf/tenant[name=\"%s\"][logical-router/segment-interface/segment=\"%s\"]"
	ResSwitch             = "/core/switch-config[name=\"%s\"]"
	ResInterfaceGroup     = "/applications/bcf/interface-group[name=\"%s\"]"
	ResMemberRuleIfaceGrp = "/applications/bcf/tenant[name=\"%s\"]/segment[name=\"%s\"]/interface-group-membership-rule[interface-group=\"%s\"][vlan=%d]"

	ResEndpoint = "/applications/bcf/info/endpoint-manager/endpoint%s" // [segment = \"%s\"][tenant = \"%s\"]"
)

// Default options used by Agent
const (
	DefBcfPort = "8443"

	PasswdEncPlainText = "plaintext"
	PasswdEncToken     = "token"
)
const (
	RespHdrTimeOutDefault     = time.Second * 30
)

type Ops interface {
	GetServer() string
	SetServer(server string)
	GetOrig() string

	GetHealth() error
	GetTenants() ([]map[string]string, error)
	CreateTenant(tName string, desc string) error
}

type BCFRestClient struct {
	server     string
	port       string
	user       string
	passwd     string
	passwdEnc  string
	id         string
	hClient    *http.Client
	pluginType string
	token      string
}

type BcfCredsConfig struct {
	Default struct {
		Ip          string `yaml:"ip"`
		AccessToken string `yaml:"access_token"`
	} `yaml:"credentials"`
}

type TenantInfo struct {
	Name        string `json:"name,omitempty"`
	Id          string `json:"id,omitempty"`
	Description string `json:"tenant-description,omitempty"`
	Origination string `json:"origination,omitempty"`
}

func IsBCFConnectivityErr(err error) bool {
	if err == ErrBCFConnTimedOut || err == ErrBCFCtrlFailOver {
		return true
	}
	return false
}

func toTenantInfo(info []byte) (TenantInfo, error) {
	var infoL []TenantInfo
	err := json.Unmarshal(info, &infoL)
	if err != nil {
		return TenantInfo{}, err
	}
	if len(infoL) <= 0 {
		return TenantInfo{}, errors.New("tenant not found")
	}
	return infoL[0], nil
}

type SegmentInfo struct {
	Name        string `json:"name,omitempty"`
	Id          string `json:"id,omitempty"`
	Description string `json:"description,omitempty"`
	Origination string `json:"origination,omitempty"`
}

func toSegmentInfo(info []byte) (SegmentInfo, error) {
	var infoL []SegmentInfo
	err := json.Unmarshal(info, &infoL)
	if err != nil {
		return SegmentInfo{}, err
	}
	if len(infoL) <= 0 {
		return SegmentInfo{}, errors.New("segment not found")
	}
	return infoL[0], nil
}


type SwitchInfo struct {
	Name        string `json:"name,omitempty"`
	MacAddr     string `json:"mac-address,omitempty"`
	FabricRole  string `json:"fabric-role,omitempty"`
	LeafGroup   string `json:"leaf-group,omitempty"`
	Shutdown    bool   `json:"shutdown,omitempty"`
	Description string `json:"description,omitempty"`
}

func toSwitchInfo(info []byte) (SwitchInfo, error) {
	var infoL []SwitchInfo
	err := json.Unmarshal(info, &infoL)
	if err != nil {
		return SwitchInfo{}, err
	}
	if len(infoL) <= 0 {
		return SwitchInfo{}, errors.New("switch not found")
	}
	return infoL[0], nil
}

type SegmentIfaceInfo struct {
	Segment string `json:"segment,omitempty"`
	Private bool   `json:"private,omitempty"`
}

func toSegmentIfaceInfo(info []byte) (SegmentIfaceInfo, error) {
	var infoL []SegmentIfaceInfo
	err := json.Unmarshal(info, &infoL)
	if err != nil {
		return SegmentIfaceInfo{}, err
	}
	if len(infoL) <= 0 {
		return SegmentIfaceInfo{}, errors.New("segment not found")
	}
	return infoL[0], nil
}

type SegmentIfaceSubnetInfo struct {
	CIDR string `json:"ip-cidr,omitempty"`
}

func toSegmentIfaceSubnetInfo(info []byte) (SegmentIfaceSubnetInfo, error) {
	var infoL []SegmentIfaceSubnetInfo
	err := json.Unmarshal(info, &infoL)
	if err != nil {
		return SegmentIfaceSubnetInfo{}, err
	}
	if len(infoL) <= 0 {
		return SegmentIfaceSubnetInfo{}, errors.New("segment subnet not found")
	}
	return infoL[0], nil
}

type TenantLR struct {
	LogicalRouter struct {
		SegmentInterface []struct {
			Segment  string `json:"segment,omitempty"`
			IpSubnet []struct {
				CIDR string `json:"ip-cidr,omitempty"`
			} `json:"ip-subnet,omitempty"`
		} `json:"segment-interface,omitempty"`
	} `json:"logical-router,omitempty"`
}

func toTenantLR(info []byte) (TenantLR, error) {
	var infoL []TenantLR
	err := json.Unmarshal(info, &infoL)
	if err != nil {
		return TenantLR{}, err
	}
	if len(infoL) <= 0 {
		return TenantLR{}, errors.New("segment logical-router not found")
	}
	return infoL[0], nil
}

type InterfaceGroupMemberInfo struct {
	Switch    string `json:"switch-name,omitempty"`
	Interface string `json:"interface-name,omitempty"`
}

type InterfaceGroupInfo struct {
	Name            string                     `json:"name,omitempty"`
	Mode            string                     `json:"mode,omitempty"`
	MemberInterface []InterfaceGroupMemberInfo `json:"member-interface,omitempty"`
	Description     string                     `json:"description,omitempty"`
}

func toInterfaceGroupInfo(info []byte) (InterfaceGroupInfo, error) {
	var infoL []InterfaceGroupInfo
	err := json.Unmarshal(info, &infoL)
	if err != nil {
		return InterfaceGroupInfo{}, err
	}
	if len(infoL) <= 0 {
		return InterfaceGroupInfo{}, errors.New("interface-group not found")
	}
	return infoL[0], nil
}

type MemberRuleIfaceGroup struct {
	InterfaceGroup string `json:"interface-group,omitempty"`
	Vlan           int    `json:"vlan,omitempty"`
}

func toMemberRuleIfaceGroupInfo(info []byte) (MemberRuleIfaceGroup, error) {
	var infoL []MemberRuleIfaceGroup
	err := json.Unmarshal(info, &infoL)
	if err != nil {
		return MemberRuleIfaceGroup{}, err
	}
	if len(infoL) <= 0 {
		return MemberRuleIfaceGroup{}, errors.New("member-rule not found")
	}
	return infoL[0], nil
}
/*
 * Copyright 2019 Big Switch Networks, Inc.
 */

package bcfrestclient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bigswitch/bcf-terraform/logger"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"time"
)

func NewFromCredsConfig(cfg *BcfCredsConfig) BCFRestClient {
	return New(cfg.Default.Ip, DefBcfPort, "", cfg.Default.AccessToken, PasswdEncToken, Orig, PluginTypeTerraform)
}

func New(server string, port string, user string, passwd string, passwdEnc string, id string, pluginType string) BCFRestClient {
	token := ""
	if passwdEnc == PasswdEncToken {
		token = passwd
	}
	e := BCFRestClient{server, port, user, passwd, passwdEnc, id, nil, pluginType, token}

	tr := &http.Transport{
		// TCP handshake timeout. KeepAlive is to use same TCP connection for http requests/responses
		DialContext: (&net.Dialer{
			Timeout:   time.Second * 10,
			KeepAlive: time.Second * 30,
		}).DialContext,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		// Max time to wait for TLS handshake
		TLSHandshakeTimeout: time.Second * 10,
		// Max time to wait for a server's first response after writing request headers if  request has "Expect: 100-continue" header
		ExpectContinueTimeout: time.Second * 10,
		// Max time to wait for a server's response headers after fully writing the request body
		ResponseHeaderTimeout: RespHdrTimeOutDefault,
		// MaxIdleConns controls the maximum number of idle (keep-alive) connections
		MaxIdleConns: 100,
		// IdleConnTimeout is the maximum time an idle (keep-alive) connection will remain idle before closing itself
		IdleConnTimeout: time.Second * 90,
	}
	hClient := &http.Client{
		Transport:     tr,
		CheckRedirect: e.handleRedirect,
	}

	e.hClient = hClient
	return e
}

func (c *BCFRestClient) GetServer() string {
	return c.server
}

func (c *BCFRestClient) SetServer(server string) {
	c.server = server
}

func (c *BCFRestClient) GetOrig() string {
	return c.id
}

func (c *BCFRestClient) setReqHeader(req *http.Request) {
	req.Header.Set("Cookie", "session_cookie="+c.token)
	req.Header.Set("Instance-ID", c.id)
	req.Header.Set("Orig-Type", c.pluginType)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
}

// performRESTWithAuth gets a token from the controller for authorization during the first request if token is not provided in BCFRestClient instance
// It uses the same token until the token expires
func (c *BCFRestClient) performRESTWithAuth(method string, resource string, data io.Reader) ([]byte, error) {
	// Initial authentication happens with the first api call
	if c.token == "" {
		err := c.authenticate()
		if err != nil {
			return nil, err
		}
	}

	bodyText, err := c.performREST(method, ResApiBase+ResApiData+resource, data)
	// Handle session expiration
	if err == ErrBCFAuth {
		err := c.authenticate()
		if err != nil {
			return nil, err
		}
		bodyText, err = c.performREST(method, ResApiBase+ResApiData+resource, data)
	}
	return bodyText, err
}

func (c *BCFRestClient) authenticate() error {
	bcf := New(c.server, DefBcfPort, c.user, c.passwd, c.passwdEnc, c.GetOrig(), c.pluginType)
	t := make(map[string]string)
	t["user"] = c.user
	t["password"] = c.passwd
	jsonStr, _ := json.Marshal(t)

	respBytes, err := bcf.performREST(MethodPost, ResApiBase+ResLogin, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}

	var resp map[string]interface{}
	json.Unmarshal(respBytes, &resp)
	if !resp["success"].(bool) {
		return errors.New(resp["error_message"].(string))
	}
	c.token = resp["session_cookie"].(string)
	return nil
}

func (c *BCFRestClient) performREST(method string, resource string, data io.Reader) ([]byte, error) {
	logger.Debugf("Server: %s, Method: %s, Resource: %s\n", c.server, method, resource)
	var URL = "https://" + c.server + ":" + c.port + resource
	req, err := http.NewRequest(method, URL, data)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	c.setReqHeader(req)

	resp, err := c.hClient.Do(req)
	if err != nil {
		// Check for Redirect error due to BCF Failover
		isRedirect, _ := regexp.MatchString("30[0-9] ", err.Error())
		if isRedirect {
			logger.Warn("BCF controller failover detected")
			return nil, ErrBCFCtrlFailOver
		}
		logger.Errorf("REST API call timed-out: %+v\n", err)
		return nil, ErrBCFConnTimedOut
	}

	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)

	if resp.Status == "401" {
		return nil, ErrBCFAuth
	}
	if (err != nil) || (resp.Status < "200") || (resp.Status > "207") {
		logger.Error("REST API call failed:", string(bodyText))
		return nil, errors.New("BCF Controller returned an error response: " + string(bodyText))
	}

	return bodyText, nil
}

func (c *BCFRestClient) handleRedirect(req *http.Request, via []*http.Request) error {
	logger.Warn("Redirect detected")
	return nil
}

func (c *BCFRestClient) GetHealth() error {
	_, err := c.performRESTWithAuth(MethodGet, ResErrors, EmptyData)
	return err
}

func (c *BCFRestClient) GetTenant(tName string) (TenantInfo, error) {
	var tInfo TenantInfo
	resource := fmt.Sprintf(ResTenant, tName)
	respBytes, err := c.performRESTWithAuth(MethodGet, resource, EmptyData)
	if err != nil {
		return tInfo, err
	}
	return toTenantInfo(respBytes)
}

func (c *BCFRestClient) CreateTenant(tName string, id string, desc string) error {
	resource := fmt.Sprintf(ResTenant, tName)
	info := TenantInfo{tName, id, desc, Orig}
	jsonStr, _ := json.Marshal(info)
	_, err := c.performRESTWithAuth(MethodPut, resource, bytes.NewBuffer(jsonStr))
	return err
}

func (c *BCFRestClient) DeleteTenant(tName string) error {
	resource := fmt.Sprintf(ResTenant, tName)
	_, err := c.performRESTWithAuth(MethodDelete, resource, EmptyData)
	return err
}

func (c *BCFRestClient) GetSegment(tName, sName string) (SegmentInfo, error) {
	resource := fmt.Sprintf(ResSegment, tName, sName)
	respBytes, err := c.performRESTWithAuth(MethodGet, resource, EmptyData)
	if err != nil {
		return SegmentInfo{}, err
	}
	return toSegmentInfo(respBytes)
}

func (c *BCFRestClient) CreateSegment(tName string, sName string, id string, desc string) error {
	resource := fmt.Sprintf(ResSegment, tName, sName)
	info := SegmentInfo{sName, id, desc, Orig}
	jsonStr, _ := json.Marshal(info)
	_, err := c.performRESTWithAuth(MethodPut, resource, bytes.NewBuffer(jsonStr))
	return err
}

func (c *BCFRestClient) DeleteSegment(tName, sName string) error {
	resource := fmt.Sprintf(ResSegment, tName, sName)
	_, err := c.performRESTWithAuth(MethodDelete, resource, EmptyData)
	return err
}

func (c *BCFRestClient) GetSegmentIface(tName string, sName string) (SegmentIfaceInfo, error) {
	resource := fmt.Sprintf(ResSegmentIface, tName, sName)
	respBytes, err := c.performRESTWithAuth(MethodGet, resource, EmptyData)
	if err != nil {
		return SegmentIfaceInfo{}, err

	}
	return toSegmentIfaceInfo(respBytes)
}

func (c *BCFRestClient) CreateSegmentIface(tName string, sName string) error {
	resource := fmt.Sprintf(ResSegmentIface, tName, sName)
	info := SegmentIfaceInfo{sName, true}
	jsonStr, _ := json.Marshal(info)
	_, err := c.performRESTWithAuth(MethodPut, resource, bytes.NewBuffer(jsonStr))
	return err
}

func (c *BCFRestClient) DeleteSegmentIface(tName, sName string) error {
	resource := fmt.Sprintf(ResSegmentIface, tName, sName)
	_, err := c.performRESTWithAuth(MethodDelete, resource, EmptyData)
	return err
}

func (c *BCFRestClient) GetSegmentIfaceSubnet(tName string, sName string, cidr string) (SegmentIfaceSubnetInfo, error) {
	resource := fmt.Sprintf(ResSegmentIfaceSubnet, tName, sName, cidr)
	respBytes, err := c.performRESTWithAuth(MethodGet, resource, EmptyData)
	if err != nil {
		return SegmentIfaceSubnetInfo{}, err
	}
	return toSegmentIfaceSubnetInfo(respBytes)
}

func (c *BCFRestClient) GetAllSubnetsForSegment(tName string, sName string) ([]string, error) {
	var cidrs []string
	resource := fmt.Sprintf(ResTenantLR, tName, sName)
	respBytes, err := c.performRESTWithAuth(MethodGet, resource, EmptyData)
	if err != nil {
		return cidrs, err
	}
	tenantLR, err := toTenantLR(respBytes)
	if err != nil {
		return cidrs, err
	}

	for _, segIf := range tenantLR.LogicalRouter.SegmentInterface {
		if segIf.Segment != sName {
			continue
		}
		for _, ipSubnet := range segIf.IpSubnet {
			cidrs = append(cidrs, ipSubnet.CIDR)
		}
	}
	return cidrs, err
}

func (c *BCFRestClient) CreateSegmentIfaceSubnet(tName string, sName string, cidr string) error {
	resource := fmt.Sprintf(ResSegmentIfaceSubnet, tName, sName, cidr)
	info := SegmentIfaceSubnetInfo{cidr}
	jsonStr, _ := json.Marshal(info)
	_, err := c.performRESTWithAuth(MethodPut, resource, bytes.NewBuffer(jsonStr))
	return err
}

func (c *BCFRestClient) DeleteSegmentIfaceSubnet(tName, sName, cidr string) error {
	resource := fmt.Sprintf(ResSegmentIfaceSubnet, tName, sName, cidr)
	_, err := c.performRESTWithAuth(MethodDelete, resource, EmptyData)
	return err
}

func (c *BCFRestClient) GetSwitch(swName string) (SwitchInfo, error) {
	resource := fmt.Sprintf(ResSwitch, swName)
	respBytes, err := c.performRESTWithAuth(MethodGet, resource, EmptyData)
	if err != nil {
		return SwitchInfo{}, err
	}
	return toSwitchInfo(respBytes)
}

func (c *BCFRestClient) CreateSwitch(swName, mac, fabricRole, leafGroup, description string, shutdown bool) error {
	resource := fmt.Sprintf(ResSwitch, swName)
	info := SwitchInfo{swName, mac, fabricRole, leafGroup, shutdown, description}
	jsonStr, _ := json.Marshal(info)
	_, err := c.performRESTWithAuth(MethodPut, resource, bytes.NewBuffer(jsonStr))
	return err
}

func (c *BCFRestClient) DeleteSwitch(swName string) error {
	resource := fmt.Sprintf(ResSwitch, swName)
	_, err := c.performRESTWithAuth(MethodDelete, resource, EmptyData)
	return err
}

func (c *BCFRestClient) GetInterfaceGroup(name string) (InterfaceGroupInfo, error) {
	resource := fmt.Sprintf(ResInterfaceGroup, name)
	respBytes, err := c.performRESTWithAuth(MethodGet, resource, EmptyData)
	if err != nil {
		return InterfaceGroupInfo{}, err
	}
	return toInterfaceGroupInfo(respBytes)
}

func (c *BCFRestClient) CreateInterfaceGroup(name string, mode string, switchIfaceMap map[string]string, description string) error {
	resource := fmt.Sprintf(ResInterfaceGroup, name)
	var switchInterfaceMembers []InterfaceGroupMemberInfo
	for swName, ifName := range switchIfaceMap {
		switchInterfaceMembers = append(switchInterfaceMembers, InterfaceGroupMemberInfo{swName, ifName})
	}
	info := InterfaceGroupInfo{name, mode, switchInterfaceMembers, description}
	jsonStr, _ := json.Marshal(info)
	_, err := c.performRESTWithAuth(MethodPut, resource, bytes.NewBuffer(jsonStr))
	return err
}

func (c *BCFRestClient) DeleteInterfaceGroup(name string) error {
	resource := fmt.Sprintf(ResInterfaceGroup, name)
	_, err := c.performRESTWithAuth(MethodDelete, resource, EmptyData)
	return err
}

func (c *BCFRestClient) GetMemberRuleIfaceGroup(tName, sName, ifaceGroupName string, vlan int) (MemberRuleIfaceGroup, error) {
	resource := fmt.Sprintf(ResMemberRuleIfaceGrp, tName, sName, ifaceGroupName, vlan)
	respBytes, err := c.performRESTWithAuth(MethodGet, resource, EmptyData)
	if err != nil {
		return MemberRuleIfaceGroup{}, err
	}
	return toMemberRuleIfaceGroupInfo(respBytes)
}

func (c *BCFRestClient) CreateMemberRuleIfaceGroup(tName, sName, ifaceGroupName string, vlan int) error {
	resource := fmt.Sprintf(ResMemberRuleIfaceGrp, tName, sName, ifaceGroupName, vlan)
	info := MemberRuleIfaceGroup{ifaceGroupName, vlan}
	jsonStr, _ := json.Marshal(info)
	_, err := c.performRESTWithAuth(MethodPut, resource, bytes.NewBuffer(jsonStr))
	return err
}

func (c *BCFRestClient) DeleteMemberRuleIfaceGroup(tName, sName, ifaceGroupName string, vlan int) error {
	resource := fmt.Sprintf(ResMemberRuleIfaceGrp, tName, sName, ifaceGroupName, vlan)
	_, err := c.performRESTWithAuth(MethodDelete, resource, EmptyData)
	return err
}

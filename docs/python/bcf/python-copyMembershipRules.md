---
# Copying Membership Rules From One Interface To Another
---

## Introduction
Use Case: Replicate membership rules in a given port(s) to another port(s)
Exact Behavior: This code will copy membership rules from one interface (or interface-group) to another interface (or interface-group)

## Usage

```bash
bash-3.2$ python copyMembershipRules.py --c 10.2.18.16 --u admin --p adminadmin --src-switch R1L1 --src-int ethernet1 --dest-switch R2L1 --dest-int ethernet2

=========== Current Membership Rules For: R1L1 - ethernet1  =====

 Tenant             Segment         VLAN
-----------------|---------------|---------
 temp3               seg1          100
 temp2               seg1          60
 temp1               seg1          50
 temp1               seg2          -1

=========== New Memership Rules For: R2L1 - ethernet2  =======

 Tenant             Segment         VLAN
-----------------|---------------|---------
 temp3               seg1          100
 temp2               seg1          60
 temp1               seg1          50
 temp1               seg2          -1
bash-3.2$
```

### Other Usage Examples

```bash
$ python copyMembershipRules.py --c 10.10.1.1 --u admin --p password --src-switch R1L1 --src-int ethernet1 --dest-ig interfaceGroup1
$ python copyMembershipRules.py --c 10.10.1.1 --u admin --p password --src-ig interfaceGroup1 --dest-switch R2L1 --dest-int ethernet3 
$ python copyMembershipRules.py --c 10.10.1.1 --u admin --p password --src-ig interfaceGroup1 --dest-ig interfaceGroup2
```

## Code Block

```python

# Big Cloud Fabric config script to copy membership rules from a source
# switch/interface (or interface-group) to a destination switch/interface
# (or interface-group)
# Example Usage:
#   python copyMembershipRules.py --c 10.2.18.16 --u admin --p adminadmin 
#     --src-switch R1L1 --src-int ethernet1 --dest-switch R2L1 --dest-int ethernet2
#   python copyMembershipRules.py --c 10.2.18.16 --u admin --p adminadmin 
#      --src-switch R1L1 --src-int ethernet1 --dest-ig interfaceGroup1
#
# For help:
#   python copyMembershipRules.py --help
#
import requests
import json
import sys
import argparse
import os

requests.packages.urllib3.disable_warnings()

### Global variables
controller_ip = ""
username = ""
password = ""
cookie = ""

def controller_request(method, path, data=""):
    if not controller_ip:
        print 'You must set controller_ip to the IP address of your BCF controller'
    controller_url = "https://%s:8443" % controller_ip

    # Append path to the controller url
    # e.g. "https://192.168.23.98:8443" + "/api/v1/auth/login"
    url = controller_url + path

    # If a cookie exists then use it in the header, otherwise create a 
    # header without a cookie
    if cookie:
        session_cookie = 'session_cookie=%s' % cookie
        headers = {"content-type": "application/json", 'Cookie': session_cookie}
    else:
        headers = {"content-type": "application/json"}

    # Submit the request
    response = requests.request(method, url, data=data, headers=headers, verify=False)
    
    # If content exists then return it, otherwise return the HTTP status code
    if response.content:
        return json.loads(response.content)
    else:
        return response.status_code

def authentication():
    global cookie
    method = 'POST'
    path = "/api/v1/auth/login"
    data = '{"user":"%s", "password":"%s"}' % (username, password)
    json_content = controller_request(method, path, data)
    cookie = json_content['session_cookie']

def authentication_revoke():
    method = "DELETE"
    path = '/api/v1/data/controller/core/aaa/session[auth-token="%s"]' % cookie
    status_code = controller_request(method, path)

def add_interface_to_segment(switch, interface, tenant, segment, vlan):
    method = 'PUT'
    path = '/api/v1/data/controller/applications/bcf/tenant[name="%s"]  \
        /segment[name="%s"]/switch-port-membership-rule[interface="%s"] \
        [switch="%s"][virtual="false"][vlan=%s]' % (tenant, segment,    \
        interface, switch, vlan)
    data = '{"switch": "%s", "interface": "%s", "vlan": "%d", "virtual": \
            "false"}' % (switch, interface, vlan)
    return controller_request(method, path, data=data)


def add_interface_group_to_segment(interface_group, tenant, segment, vlan):
    method = 'PUT'
    path = '/api/v1/data/controller/applications/bcf/tenant[name="%s"]/segment \
           [name="%s"]/interface-group-membership-rule[vlan=%s]  \
           [interface-group="%s"]' %(tenant, segment, vlan, interface_group)
    data = '{"vlan": %s, "interface-group": "%s"}' %(vlan, interface_group)
    return controller_request(method, path, data=data)


def get_member_rules_switchInterface(switch, interface):
        # Get member rules for a switch interface. Returns list of tenants 
        # and segments
        method = 'GET'
        path = '/api/v1/data/controller/applications/bcf/info/endpoint-manager \
               /member-rule[interface="%s"][switch="%s"]' % (interface, switch)
        tenantSegmentVlanDict= dict()
        for line in controller_request(method, path):
                if line['tenant'] not in tenantSegmentVlanDict:
                    tenantSegmentVlanDict[line['tenant']] = dict()
                    if {line['vlan'] != 'untagged'}:
                        tenantSegmentVlanDict[line['tenant']][line['segment']] = line['vlan']
                    else:
                        tenantSegmentVlanDict[line['tenant']][line['segment']] = line['-1']
                else:
                    if {line['vlan'] != 'untagged'}:
                        tenantSegmentVlanDict[line['tenant']][line['segment']] = line['vlan']
                    else:
                        tenantSegmentVlanDict[line['tenant']][line['segment']] = line['-1']

        return tenantSegmentVlanDict


def get_member_rules_interface_group(interface_group):
        # Get member rules for an interface group. Returns list of tenants
        # and segments
        method = 'GET'
        path = '/api/v1/data/controller/applications/bcf/info/endpoint-manager \
               /member-rule[interface-group="%s"]' % (interface_group)
        tenantSegmentVlanDict= dict()
        for line in controller_request(method, path):
                if line['tenant'] not in tenantSegmentVlanDict:
                    tenantSegmentVlanDict[line['tenant']] = dict()
                    if {line['vlan'] != 'untagged'}:
                        tenantSegmentVlanDict[line['tenant']][line['segment']] = line['vlan']
                    else:
                        tenantSegmentVlanDict[line['tenant']][line['segment']] = line['-1']
                else:
                    if {line['vlan'] != 'untagged'}:
                        tenantSegmentVlanDict[line['tenant']][line['segment']] = line['vlan']
                    else:
                        tenantSegmentVlanDict[line['tenant']][line['segment']] = line['-1']

        return tenantSegmentVlanDict



def main(cntr_ip, uname, pwrd, Source_Switch, Source_Interface,  \
    Source_Interface_Group, Dest_Switch, Dest_Interface, Dest_Interface_Group):
    global controller_ip
    global username
    global password
    controller_ip = cntr_ip
    username = uname
    password = pwrd

    if not Source_Switch and not Source_Interface and Source_Interface_Group:
            print "\n======================== Current Membership Rules For: %s \
                   ================\n" % (Source_Interface_Group)
    elif Source_Switch and Source_Interface and not Source_Interface_Group:
            print "\n======================== Current Membership Rules For: %s \
                   - %s  ================\n" % (Source_Switch, Source_Interface)
    else:
            print "Error: Please Check the Input Parameters. You can only  \
                define SourceSwitch/Interface or Source_Interface_Group"
            sys.exit()

    authentication()
    if not cookie:
            print "Error: Authentication Failure. Please check the credentials"
            sys.exit()
    if (Source_Switch and Source_Interface):
        tenantSegmentVlanDict = get_member_rules_switchInterface(Source_Switch,\
                                   Source_Interface)
    else:
        tenantSegmentVlanDict = get_member_rules_interface_group(Source_Interface_Group)

    tenantSegmentVlanList = []
    print " Tenant             Segment         VLAN\n---------------|---------------|---------"
    for tenant in tenantSegmentVlanDict:
            for segment in tenantSegmentVlanDict[tenant]:
                    print " %s\t\t%s\t\t%d" % (tenant, segment,  \
                          tenantSegmentVlanDict[tenant][segment])
                    tenantSegmentVlanList.append((tenant,segment, \
                          tenantSegmentVlanDict[tenant][segment]))

    print "\n------ Return Code/Reason ------"
    if(Dest_Switch and Dest_Interface):
        for tenantsegmentvlan in tenantSegmentVlanList:
                print add_interface_to_segment(Dest_Switch, Dest_Interface,  \
                tenantsegmentvlan[0], tenantsegmentvlan[1], tenantsegmentvlan[2])
    else:
        for tenantsegmentvlan in tenantSegmentVlanList:
                print add_interface_group_to_segment(Dest_Interface_Group, \
                tenantsegmentvlan[0], tenantsegmentvlan[1], tenantsegmentvlan[2])


    if (Dest_Switch and Dest_Interface):
        print "\n======================== New Memership Rules For: %s - %s  \
               ================\n" % (Dest_Switch, Dest_Interface)
        tenantSegmentVlanDict = get_member_rules_switchInterface(Dest_Switch, Dest_Interface)
    else:
        print "\n======================== New Membership Rules For: %s \
               ================\n" % (Dest_Interface_Group)
        tenantSegmentVlanDict = get_member_rules_interface_group(Dest_Interface_Group)

    tenantSegmentVlanList = []
    print " Tenant             Segment         VLAN\n---------------|---------------|---------"
    for tenant in tenantSegmentVlanDict:
            for segment in tenantSegmentVlanDict[tenant]:
                    print " %s\t\t%s\t\t%d" % (tenant, segment, tenantSegmentVlanDict[tenant][segment])
                    tenantSegmentVlanList.append((tenant,segment,tenantSegmentVlanDict[tenant][segment]))

    authentication_revoke()

if __name__ == '__main__':

    descr = """
    This utility copies membership rules from a Source switch/interface \
     (or interface-group) to a Destination switch/interface (or interface-group)
    """
    parser = argparse.ArgumentParser(prog=os.path.basename(__file__),
            formatter_class=argparse.RawDescriptionHelpFormatter, \
                                    description=descr)
    parser.add_argument('--c', help='Controller IP address', required=True)
    parser.add_argument('--u', help='User Name', required=True)
    parser.add_argument('--p', help='Password', required=True)
    parser.add_argument('--src-switch', help='Source Switch')
    parser.add_argument('--src-int', help='Source Interface')
    parser.add_argument('--src-ig', help='Source Interface Group')

    parser.add_argument('--dest-switch', help='Destination Switch')
    parser.add_argument('--dest-int', help='Destination Interface')
    parser.add_argument('--dest-ig', help='Destination Interface Group')


    args = parser.parse_args()

    if (args.src_switch and args.src_int and not args.src_ig):
            pass
    elif (not args.src_switch and not args.src_int and args.src_ig):
            pass
    else:
            parser.error("Invalid Arguments: Please Define [--src-switch & --src-int] OR [--src-ig]")

    if (args.dest_switch and args.dest_int and not args.dest_ig):
            pass
    elif (not args.dest_switch and not args.dest_int and args.dest_ig):
            pass
    else:
            parser.error("Invalid Arguments: Please Define [--dest-switch & --dest-int] OR [--dest-ig]")

    main(args.c, args.u, args.p, args.src_switch, args.src_int, args.src_ig, \
         args.dest_switch, args.dest_int, args.dest_ig)
```

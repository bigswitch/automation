#!/usr/bin/env python
# Big Monitoring Fabric Out-of-band config script 
import requests
import json 
import sys

requests.packages.urllib3.disable_warnings()

DRY_RUN = False

controller_ip = ""
username = ""
password = ""

# do not modify
# do not set 
cookie = ""
bigtap_path = '/api/v1/data/controller/applications/bigtap/'
controller_path = '/api/v1/data/controller/core/'

def controller_request(method, path, data="", dry_run=False):
    if not controller_ip:
        print 'You must set controller_ip to the IP address of your controller'
    controller_url = "https://%s:443" % controller_ip
    # append path to the controller url, e.g. "https://192.168.23.98:8443" + "/api/v1/auth/login"
    url = controller_url + path
    # if a cookie exists then use it in the header, otherwise create a header without a cookie
    if cookie:
        session_cookie = 'session_cookie=%s' % cookie
        headers = {"content-type": "application/json", 'Cookie': session_cookie}
    else:
        headers = {"content-type": "application/json"}
    if dry_run:
        print 'METHOD ', method, ' URL ', url, ' DATA ', data, ' HEADERS ', headers
        return None
    else:
        # submit the request 
        response = requests.request(method, url, data=data, headers=headers, verify=False)
        # if content exists then return it, otherwise return the HTTP status code
        if response.content:
            return json.loads(response.content)
        else:
            return response.status_code
                        
def make_request(message, verb, path, data):
    result = controller_request(verb, path, data=data, dry_run=DRY_RUN)
    if result == 200 or DRY_RUN:
        print message + ' ... ok'
        return True
    print result
    sys.exit(1)

def authentication():
    global cookie
    method = 'POST'
    path = "/auth/login"
    data = '{"user":"%s", "password":"%s"}' % (username, password)
    json_content = controller_request(method, path, data)
    cookie = json_content['session_cookie']
    print 'Login to %s successful' %controller_ip

def authentication_revoke():
    method = "DELETE"
    path = '/api/v1/data/controller/core/aaa/session[auth-token="%s"]' % cookie
    status_code = controller_request(method, path)
    if status_code == 200:
        print 'Logout successful'

def get_controller_version():
    method = 'GET'
    path = '/rest/v1/system/version'
    data = '{}'
    json_content = controller_request(method, path, data=data, dry_run=True)
    return json_content[0] if type(json_content) == list else None

def configure_bigtap_interface_role(switch_dpid, interface, name, role):
    method = 'PUT'
    path = bigtap_path+ 'interface-config[interface="%s"][switch="%s"]' % (interface, switch_dpid)
    data = '{"interface": "%s", "switch": "%s", "role": "%s", "name": "%s"}' % (interface, switch_dpid, role, name)
    print 'Assign bigtap role to interface %s.' % interface, controller_request(method, path, data=data, dry_run=False)

def add_policy(specs):
    try:
        name = specs['name']
        action = specs['action']
        priority = specs['priority']
        duration = specs['duration']
        start_time = specs['start_time']
        delivery_packet_count = specs['delivery_packet_count']
        interfaces = specs['interfaces']
        rules = specs['rules']
    except KeyError, e:
        print "policy specs error %s" % str(e)
        sys.exit(1)

    path = bigtap_path+ 'view[name="admin-view"]/policy[name="%s"]' % name
    data = '{"name": "%s"}' % name
    make_request('1. Create policy', 'PUT', path, data)

    #def set_policy_action(name, action):
    path = bigtap_path+ 'view[name="admin-view"]/policy[name="%s"]' % name
    data = '{"action": "%s"}' % action
    make_request('2. Set policy action', 'PATCH', path, data)
    
    #def set_policy_priority(name, priority):
    path = bigtap_path+ 'view[name="admin-view"]/policy[name="%s"]' % name
    data = '{"priority": %s}' % priority
    make_request('3. Set priority', 'PATCH', path, data)

    #def start_policy(name, start_time=1477677396):
    duration = 0
    delivery_packet_count = 0
    path = bigtap_path+ 'view[name="admin-view"]/policy[name="%s"]' % name
    data = '{"duration": %s, "start-time": %s, "delivery-packet-count": %s}' % (duration, start_time, delivery_packet_count)
    make_request( '4. Set policy start time, duration, delivery pkt count', 'PATCH', path, data)

    index = 5
    for interface in interfaces:
        role = interfaces[interface]
        path = bigtap_path+ 'view[name="admin-view"]/policy[name="%s"]/%s-group[name="%s"]' % (name, role, interface)
        data = '{"name": "%s"}'  % interface
        make_request(str(index)+'. Add interface %s as %s' %(interface, role), 'PUT', path, data)
        index += 1

    for rule in rules:
        path = bigtap_path+ 'view[name="admin-view"]/policy[name="%s"]/rule[sequence=%s]' % (name, rule['sequence'])
        data = json.dumps(rule)
        make_request(str(index)+'. Add rule %s' % str(rule), 'PUT', path, data)
        index += 1

def create_tunnel(specs, dry_run=False):
    """ """
    try:
        switch_dpid = specs['switch_dpid']
        tunnel_name = specs['tunnel_name']
        destination_ip = specs['destination_ip']
        source_ip = specs['source_ip']
        mask = specs['mask']
        gateway_ip = specs['gateway_ip']
        tunnel_src_ip = specs['destination_ip']
        vpn_key = specs['vpn_key']
        encap_type = specs['encap_type']
        interface = specs['interface']
        direction = specs['direction']
        loopback_interface = ''
        if direction == 'bidirectional' or direction == 'transmit-only':
            loopback_interface = specs['loopback_interface']
    except KeyError, e:
        print "tunnel specs error %s" % str(e)
        sys.exit(1)

    #def create_tunnel_interface(switch_dpid, tunnel_name):
    path = controller_path+ 'switch[dpid="%s"]/interface[name="%s"]' % (switch_dpid, tunnel_name)
    data = '{"name": "%s"}' % tunnel_name
    make_request('1. Create GRE tunnel %s' % tunnel_name, 'PUT', path, data)

    # set_tunnel_destination(switch_dpid, tunnel_name, destination_ip):
    path = controller_path+ 'switch[dpid="%s"]/interface[name="%s"]/ip-config' % (switch_dpid, tunnel_name)
    data = '{"destination-ip": "%s"}' % destination_ip
    make_request('2. Add IP destination to tunnel %s' % destination_ip, 'PUT', path, data)

    #def set_tunnel_encap_type(switch_dpid, tunnel_name, encap_type, vpn_key = 0):
    path = controller_path+ 'switch[dpid="%s"]/interface[name="%s"]' % (switch_dpid, tunnel_name)
    data = '{"vpn-key": %s, "encap-type": "%s"}' % (vpn_key, encap_type)
    make_request('3. Set tunnel key & encapsulation type', 'PATCH', path, data)

    #def set_tunnel_parent_interface(switch_dpid, tunnel_name, interface):
    path = controller_path+ 'switch[dpid="%s"]/interface[name="%s"]' % (switch_dpid, tunnel_name)
    data = '{"parent-interface": "%s"}' % interface
    print '4. Set switch interface', controller_request('PATCH', path, data=data, dry_run=dry_run)
    make_request('4. Set switch interface', 'PATCH', path, data)

    #def set_tunnel_direction(switch_dpid, tunnel_name, direction, loopback_interface = ""):
    #""" bidirectional  receive-only   transmit-only """
    # {"direction": "receive-only", "loopback-interface": "", "type": "tunnel"}
    path = controller_path+ 'switch[dpid="%s"]/interface[name="%s"]' % (switch_dpid, tunnel_name)
    data = '{"direction": "%s", "loopback-interface": "%s", "type": "tunnel"}' % (direction, loopback_interface)
    make_request('5. Set tunnel direction and loopback interface (in case of transmit)', 'PATCH', path, data)

    #def set_tunnel_direction(switch_dpid, tunnel_name, tunnel_source_ip, mask, gateway_ip):
    path = controller_path+ 'switch[dpid="%s"]/interface[name="%s"]/ip-config' % (switch_dpid, tunnel_name)
    data = '{"source-ip": "%s", "ip-mask": "%s", "gateway-ip": "%s"}' % (source_ip, mask, gateway_ip)
    make_request('6. Set IP source of tunnel %s' % source_ip, 'PATCH', path, data)

if __name__ == '__main__':
    switch_dpid = '00:00:70:72:cf:c1:7c:7b'
    tunnel_name = 'tunnel-vmware0'
    interface = 'ethernet41'
    authentication()

    tunnel_specs = {
        'switch_dpid': switch_dpid, # Filter-1
        'tunnel_name': tunnel_name,
        'destination_ip': '10.2.19.125',
        'source_ip': '10.10.9.242', # switch uses this IP address for the tunnel interface
        'mask': '255.255.255.240',
        'gateway_ip': '10.10.9.241',
        'vpn_key': 0,
        'encap_type': 'gre',
        'interface': interface, # tunnel interface connected to mgmt switch
        'direction': 'receive-only',
        'loopback_interface': ''
    }
    create_tunnel(tunnel_specs, dry_run=False)

    configure_bigtap_interface_role(switch_dpid, tunnel_name, tunnel_name+'-Filter', 'filter')

    policy_specs = {
        'name': 'vmware0',
        'action': 'forward',
        'priority': 100,
        'duration': 0,
        'start_time': 0,
        'delivery_packet_count': 0,
        'interfaces': {
            tunnel_name+'-Filter': 'filter',
            'Bro-IDS-H1-eth3': 'delivery'
        },
        'rules': [
            {"any-traffic": True, "sequence": 1}
        ]
    }
    add_policy(policy_specs)

    authentication_revoke()
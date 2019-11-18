#!/usr/bin/python

import os
import sys
### Set the path to your modules directory here
sys.path.append('/root/ansible/lib/ansible/modules')

from ansible.module_utils.basic import *
from ansible.module_utils.network.bigswitch.bigswitch import *
from ansible.module_utils.pycompat24 import get_exception


def diff(existing_tenants, module):
    """
    1. fill canonical tenant logical router structure with module input
    2. compare input structure with matching structure, keeping differing fields
    3. return the diff and let tenant apply only the diff
    """
    matching = [tenant for tenant in existing_tenants if tenant['name'] == module.params['name']]
    if matching:
        return True
    return False

def bcf_tenant(module):
    try:
        access_token = module.params['access_token'] or os.environ['BIGSWITCH_ACCESS_TOKEN']
    except KeyError:
        e = get_exception()
        module.fail_json(msg='Unable to load %s' % e.message )

    name = module.params['name']
    logical_router_interfaces = module.params['logical_router_interfaces']
    logical_router_tenant_interfaces = module.params['logical_router_tenant_interfaces']
    routes = module.params['routes']
    next_hop_groups = module.params['next_hop_groups']
    policy_lists = module.params['policy_lists']
    #inbound_policy = module.params['inbound_policy']
    tenant_id = module.params['tenant_id']
    state = module.params['state']
    controller = module.params['controller']

    rest = Rest(module,
                {'Content-type': 'application/json',
                 'Cookie': 'session_cookie='+access_token},
                'https://'+controller+':8443/api/v1/data/controller/applications/bcf')

    if None in (name, state, controller):
        module.fail_json(msg='parameter `name` is missing')

    response = rest.get('tenant?config=true', data={})
    if response.status_code != 200:
        module.fail_json(msg="failed to obtain existing tenant config: {}".format(response.json['description']))

    #response = rest.get('tenant?select=logical-router', data={})
    #if response.status_code != 200:
    #    module.fail_json(msg="failed to obtain existing tenant {} logical router config: {}".format(name, response.json['description']))

    config_present = False
    if diff(response.json, module):
        config_present = True

    if state in ('present') and config_present:
        module.exit_json(changed=False)

    if state in ('absent') and not config_present:
        module.exit_json(changed=False)

    if state in ('present'):
        # TODO: implement tenant config matching
        if not config_present:
            response = rest.put('tenant[name="%s"]' % name, data={'name': name})
            if response.status_code != 204:
                module.fail_json(msg="error creating tenant '{}': {}".format(name, response.info['msg']))

        for logical_router_interface in logical_router_interfaces:
            if logical_router_interfaces[logical_router_interface]:
                subnet = logical_router_interfaces[logical_router_interface]
            else:
                subnet = None

            data = {"segment": logical_router_interface}
            path = 'tenant[name="%s"]/logical-router/segment-interface[segment="%s"]' %( name, logical_router_interface )
            response = rest.put(path, data=data)
            if response.status_code != 204:
                module.fail_json(msg="error adding segment interface to router '{}': {}".format(name, response.info))

            if subnet:
                data = {"ip-cidr": subnet}
                path += '/ip-subnet[ip-cidr="\'%s\'"]' % subnet
                response = rest.put(path, data=data)
                if response.status_code != 204:
                    module.fail_json(msg="error configuring ip subnet to router interface '{}': {}".format(name, response.info))

        for remote_tenant in logical_router_tenant_interfaces:
            data = {"remote-tenant": remote_tenant}
            path = 'tenant[name="%s"]/logical-router/tenant-interface[remote-tenant="%s"]' %( name, remote_tenant )
            response = rest.put(path, data=data)
            if response.status_code != 204:
                module.fail_json(msg="error adding system tenant interface to router '{}': {}".format(name, response.info))

        for destination_subnet in routes:
            try:
                next_hop_type, next_hop_value = routes[destination_subnet].split(':')
            except ValueError:
                module.fail_json(msg="malformed route to '{}' for tenant '{}': {}".format(destination_subnet, name, response.info))
#            data = {'next-hop': {next_hop_type: next_hop_value}, 'dst-ip-subnet': destination_subnet}
            data = {'next-hop': {next_hop_type: next_hop_value}, 'dst-ip-subnet': destination_subnet, 'preference': 1}
            path = 'tenant[name="%s"]/logical-router/static-route[dst-ip-subnet="\'%s\'"]' %( name, destination_subnet)
#            path = 'tenant[name="%s"]/logical-router/static-route[dst-ip-subnet="\'%s\'"][preference=1]' %( name, destination_subnet)
            response = rest.put(path, data=data)
            if response.status_code != 204:
                module.fail_json(msg="error configuring route for destination subnet '{}': {}".format(path, response.info))

        for next_hop_group in next_hop_groups:
            ips = next_hop_groups[next_hop_group]

            data = {'name': next_hop_group}
            path = 'tenant[name="%s"]/logical-router/next-hop-group[name="%s"]' %( name, next_hop_group)
            response = rest.put(path, data=data)
            if response.status_code != 204:
                module.fail_json(msg="error configuring next-hop-group '{}': {}".format(next_hop_group, response.info))

            for ip in ips:
                data = {'ip-address': ip}
#                       'tenant[name="%s"]/logical-router/next-hop-group[name="%s"]/ip-address[ip-address="10.0.5.2"] {"ip-address": "10.0.5.2"}
                path = 'tenant[name="%s"]/logical-router/next-hop-group[name="%s"]/ip-address[ip-address="\'%s\'"]' %( name, next_hop_group, ip)
                response = rest.put(path, data=data)
                if response.status_code != 204:
                    module.fail_json(msg="error configuring next-hop-group '{}': {}".format(next_hop_group, response.info))

        for policy_list in policy_lists:
            rules = policy_lists[policy_list]
            if not isinstance(rules, list):
                module.fail_json(msg="policy list '{}' must be a list of rules: {}".format(policy_list, response.info))
            data = {'name': policy_list}
            path = 'tenant[name="%s"]/logical-router/policy-list[name="%s"]' % (name, policy_list)
            response = rest.put(path, data=data)
            if response.status_code != 204:
                module.fail_json(msg="error configuring policy list '{}': {}".format(policy_list, response.info))
            for rule in rules:
                data = rule
                if data['action'] == 'next-hop':
                    path = 'tenant[name="%s"]/logical-router/policy-list[name="%s"]\
                            /rule[next-hop/next-hop-group="%s"][seq=%s][segment-interface="%s"]\
                            [dst/segment="%s"][dst/tenant="%s"][action="%s"]'\
                                             % (name,
                                              policy_list,
                                              data["next-hop"]["next-hop-group"],
                                              data["seq"],
                                              data.get("segment-interface", ""),
                                              data["dst"].get("segment", ""),
                                              data["dst"].get("tenant", ""),
                                              data["action"])
                else:
                    path = 'tenant[name="%s"]/logical-router/policy-list[name="%s"]\
                            /rule[seq=%s][action="%s"]'\
                                             % (name, policy_list, data["seq"], data["action"])

                response = rest.put(path, data=data)
                if response.status_code != 204:
                    module.fail_json(msg="error configuring policy list '{}': {}".format(policy_list, response.info))

#        if inbound_policy:
#            data = {'inbound-policy': inbound_policy}
#            path = 'tenant[name="%s"]/logical-router' % name
#            response = rest.patch(path, data=data)
#            if response.status_code != 204:
#                module.fail_json(msg="error applying policy list to tenant '{}': {}".format(name, response.info))

        module.exit_json(changed=True)

    if state in ('absent'):
        response = rest.delete('tenant[name="%s"]' % name, data={})
        if response.status_code == 204:
            module.exit_json(changed=True)
        else:
            module.fail_json(msg="error deleting tenant '{}': {}".format(name, response.info['msg']))



#def bcf_tenant(module):
#    try:
#        access_token = module.params['access_token'] or os.environ['BIGSWITCH_ACCESS_TOKEN']
#    except KeyError:
#        e = get_exception()
#        module.fail_json(msg='Unable to load %s' % e.message )
#
#    name = module.params['name']
#    logical_router_interfaces = module.params['logical_router_interfaces']
#    logical_router_tenant_interfaces = module.params['logical_router_tenant_interfaces']
#    routes = module.params['routes']
#    next_hop_groups = module.params['next_hop_groups']
#    policy_lists = module.params['policy_lists']
##    inbound_policy = module.params['inbound_policy']
#    tenant_id = module.params['tenant_id']
#    state = module.params['state']
#    controller = module.params['controller']
#
#    rest = Rest(module,
#                {'Content-type': 'application/json',
#                 'Cookie': 'session_cookie='+access_token},
#                'https://'+controller+':8443/api/v1/data/controller/applications/bcf')
#
#    if None in (name, state, controller):
#        module.fail_json(msg='parameter `name` is missing')
#
#    response = rest.get('/info/endpoint-manager/tenant', data={'name':name})
#
#    if response.status_code != 200:
#        module.fail_json(msg="failed to obtain existing tenant config: {}".format(response.json['description']))
#
#    config_present = False
#    matching = [tenant for tenant in response.json if tenant['name'] == name]
#    if matching:
#        config_present = True
#
#    if state in ('present') and config_present:
#        module.exit_json(changed=False)
#
#    if state in ('absent') and not config_present:
#        module.exit_json(changed=False)
#        
#    if state in ('present'):
#	response = rest.put('/tenant[name="%s"]' % name,  data={"name": name})
#        if response.status_code == 204:
#            module.exit_json(changed=True)
#        else:
#            module.fail_json(msg="error creating tenant '{}': {}".format(name, response.json['description']))
#
#    if state in ('absent'):
#	response = rest.delete('tenant[name="%s"]' % name, data={})
#        if response.status_code == 204:
#            module.exit_json(changed=True)
#        else:
#            module.fail_json(msg="error deleting tenant '{}': {}".format(name, response.json['description']))
#
#def main():
#    module = AnsibleModule(
#        argument_spec=dict(
#            name=dict(type='str', required=True),
#            controller=dict(type='str', required=True),
#            state=dict(choices=['present', 'absent'], default='present'),
#            validate_certs=dict(type='bool', default='False'),
#            access_token=dict(aliases=['BIGSWITCH_ACCESS_TOKEN'], no_log=True)
#        )
#    )
#
#    try:
#        tenant(module)
#    except Exception:
#        e = get_exception()
#        module.fail_json(msg=str(e))
#
#if __name__ == '__main__':
#    main()

def main():
    module = AnsibleModule(
        argument_spec=dict(
            name=dict(type='str', required=True),
            tenant_id=dict(type='str', required=False),
            logical_router_interfaces=dict(type='dict', default={}),
            logical_router_tenant_interfaces=dict(type='list', default=[]),
            routes=dict(type='dict', default={}),
            next_hop_groups=dict(type='dict', default={}),
            policy_lists=dict(type='dict', default={}),
           #inbound_policy=dict(type='str', required=False),
    #        inbound_policy=dict(type='dict', default={}),
            controller=dict(type='str', required=True),
            state=dict(choices=['present', 'absent'], default='present'),
            validate_certs=dict(type='bool', default='False'), # TO DO: change this to default = True
            access_token=dict(type='str', no_log=True)
        )
    )

    try:
        bcf_tenant(module)
    except Exception:
        e = get_exception()
        module.fail_json(msg=str(e))

if __name__ == '__main__':
    main()

#!/usr/bin/python
# -*- coding: utf-8 -*-

# Ansible module to manage Big Cloud Fabric tenant configurations
# (c) 2016, Don Jayakody <don@bigswitch.com>,
#
# This file is part of Ansible
#
# Ansible is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# Ansible is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with Ansible.  If not, see <http://www.gnu.org/licenses/>.

ANSIBLE_METADATA = {'status': ['preview'],
                    'supported_by': 'community',
                    'version': '1.0'}

DOCUMENTATION = '''
---
module: bcf_tenant
short_description: Create and remove a bigmon inline service chain.
description:
    - Create and remove a bigmon inline service chain.
version_added: "2.3"
options:
  name:
    description:
     - The name of the chain.
    required: true
  state:
    description:
     - Whether the service chain should be present or absent.
    default: present
    choices: ['present', 'absent']
  controller:
    description:
     - The controller IP address.
    required: true
  validate_certs:
    description:
      - If C(false), SSL certificates will not be validated. This should only be used
        on personally controlled devices using self-signed certificates.
    required: false
    default: true
    choices: [true, false]
  access_token:
    description:
     - Bigmon access token.
    required: false

notes:
  - An environment variable can be used, BIGSWITCH_ACCESS_TOKEN.
'''


EXAMPLES = '''
- name: bigmon inline service chain
      bigmon_chain:
        name: MyChain
        controller: '{{ inventory_hostname }}'
        state: present
'''


RETURN = '''
{
    "changed": true,
    "invocation": {
        "module_args": {
            "access_token": null,
            "controller": "192.168.86.221",
            "name": "MyChain",
            "state": "present",
            "validate_certs": false
        },
        "module_name": "bigmon_chain"
    }
}
'''

import os
import sys
sys.path.append('/var/lib/awx/projects/_11__github/modules')

from ansible.module_utils.basic import *
from bigswitch_utils import Rest, Response
from ansible.module_utils.pycompat24 import get_exception

def bcfTenant(module):
    try:
        access_token = module.params['access_token'] or os.environ['BIGSWITCH_ACCESS_TOKEN']
    except KeyError:
        e = get_exception()
        module.fail_json(msg='Unable to load %s' % e.message )

    segIntState = module.params['segIntState']
    controller = module.params['controller']
    tenantName = module.params['tenantName']
    segmentName = module.params['segmentName']
    ipcidr = module.params['ipcidr']

    rest = Rest(module,
                {'content-type': 'application/json', 'Cookie': 'session_cookie='+access_token},
                'https://'+controller+':8443/api/v1/data/controller/applications/bcf')

    if None in (tenantName, segmentName, segIntState, controller):
        module.fail_json(msg='parameter `name` is missing')

    response = rest.get('/info/endpoint-manager/tenant', data={'name':tenantName})

    if response.status_code != 200:
        module.fail_json(msg="failed to obtain existing tenant config: {}".format(response.json['description']))

    config_present = False
    matching = [tenant for tenant in response.json if tenant['name'] == tenantName]
    if matching:
        config_present = True
	
    if segIntState in ('present') and config_present:
	    # Create logical router
	    response = rest.put('/tenant[name="%s"]/logical-router' % tenantName, data= {})
            if response.status_code != 204:
        	module.fail_json(msg="failed to create logical router for the tenant config: {}".format(response.json['description']))
		
	    # Create segment interface
	    response = rest.put('/tenant[name="%s"]/logical-router/segment-interface[segment="%s"]' % (tenantName, segmentName), 
			data= {"segment": segmentName})
	    if response.status_code != 204:
        	module.fail_json(msg="failed to create segment interface for the logical router: {}".format(response.json['description']))

	    # Add interace segment IP
	    response = rest.put('/tenant[name="%s"]/logical-router/segment-interface[segment="%s"] \
				/ip-subnet[no-autoconfig="false"][withdraw="false"][ip-cidr="%s"]' % (tenantName, segmentName,ipcidr),
				data = {"no-autoconfig": 'false',  "ip-cidr": ipcidr,  "withdraw": 'false'})
	    if response.status_code != 204:
        	module.fail_json(msg="failed to create segment interface ip for the logical router: {}".format(response.json['description']))
            
	    module.exit_json(changed=True)

    if segIntState in ('absent') and config_present:
	    # Delete segment interface
	    response = rest.delete('/tenant[name="%s"]/logical-router/segment-interface' % tenantName,
			data= {"segment": segmentName})
	    if response.status_code != 204:
        	module.fail_json(msg="failed to delete segment interface for the logical router: {}".format(response.json['description']))

	    module.exit_json(changed=True)

    if tenantState in ('absent') and not config_present:
        module.exit_json(changed=False)


    module.exit_json(changed=False)
        
def main():
    module = AnsibleModule(
        argument_spec=dict(
            tenantName=dict(type='str', required=True),
            segmentName=dict(type='str', required=True),
            ipcidr=dict(type='str', required=True),
            controller=dict(type='str', required=True),
            segIntState=dict(choices=['present', 'absent'], default='present'),
            validate_certs=dict(type='bool', default='False'),
            access_token=dict(aliases=['BIGSWITCH_ACCESS_TOKEN'], no_log=True)
        )
    )

    try:
        bcfTenant(module)
    except Exception:
        e = get_exception()
        module.fail_json(msg=str(e))

if __name__ == '__main__':
    main()

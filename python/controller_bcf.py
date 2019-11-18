#
# Simple BCF config script 
# No error checking 
#
import requests
import json 
import sys

requests.packages.urllib3.disable_warnings()

class Controller(object):
    """

    controller version 4.x

    """
    def __init__(self, controller_ip, access_token):
        self.bcf_path = '/api/v1/data/controller/applications/bcf'
        self.core_path = '/api/v1/data/controller/core'
        self.controller_ip = controller_ip
        self.access_token = access_token

    def controller_request(self, method, path, data="", dry_run=False):
        if not self.controller_ip:
            print( 'You must set controller_ip to the IP address of your controller' )
        controller_url = "https://%s:8443" % self.controller_ip
        # append path to the controller url, e.g. "https://192.168.23.98:8443" + "/api/v1/auth/login"
        url = controller_url + path
        # if a cookie exists then use it in the header, otherwise create a header without a cookie
        session_cookie = 'session_cookie=%s' % self.access_token
        headers = {"content-type": "application/json", 'Cookie': session_cookie}

        if dry_run:
            print( 'METHOD ', method, ' URL ', url, ' DATA ', data, ' HEADERS ', headers )
            return None
        else:
            # submit the request
            response = requests.request(method, url, data=data, headers=headers, verify=False)
            # if content exists then return it, otherwise return the HTTP status code
            #if response.content:
            #    return response.content
            #else:
            return response

    def make_request(self, verb, path, data, core_path = False):
        if core_path:
            return self.controller_request(verb, self.core_path + path, data=data, dry_run=False)
        return self.controller_request(verb, self.bcf_path + path, data=data, dry_run=False)
        
    def interface_group(self, name, mode='', origination='', action='add'):
        path = '/interface-group[name="%s"]' % name
        if origination and mode:
            data = '{"name": "%s", "mode": "%s", "origination": "%s"}' % (name, mode, origination)
        elif origination:
            data = '{"name": "%s", "origination": "%s"}' % (name, origination)
        else:
            data = '{"name": "%s"}' % name
        if action == 'add':
            return self.make_request('PUT', path, data=data)
        elif action == 'delete':
            return self.make_request('DELETE', path, data=data)
        elif action == 'get':
            response = self.make_request('GET', path, data=data).json()
            if response:
                return response[0] if response else []

    def interface_groups(self, interface_groups=[], action='add'):
        """ add each interface_group in list of interface_groups using the function add_interface_group """
        if action == 'add':
            for interface_group in interface_groups:
                interface_group(interface_group, action=action)
        elif action == 'get':
            path = '/info/fabric/interface-group/detail'
            response = self.make_request('GET', path, data='{}').json()
            return response
            if response:
                return response[0] if response else []
        
    def interface_group_member(self, switch, interface, interface_group, action='add'):
        path = '/interface-group[name="%s"]/member-interface[switch-name="%s"][interface-name="%s"]'% (interface_group, switch, interface)
        data = '{"switch-name": "%s", "interface-name": "%s"}' % (switch, interface)
        return self.make_request('PUT' if action == 'add' else 'DELETE', path, data=data)
    
    def interface_group_members(self, switch_interface, interface_group, action='add'):
        """ add each (switch, interface) pair in interfaces to interface-group interface_group  """
        for (switch, interface) in switch_interface:
            self.interface_group_member(switch, interface, interface_group, action=action)
    
    def tenant(self, name, origination='', action='add'):
        path = '/tenant[name="%s"]?select=name&single=true' % name
        response = self.make_request('GET', path, data='{}')
        config_present = True if response.status_code != 404 else False

        path = '/tenant[name="%s"]' % name
        data = '{"name": "%s"}' % name
        if origination:
            data = '{"name": "%s", "origination": "%s"}' % (name, origination)
        if action == 'add' and not config_present:
            return self.make_request('PUT', path, data=data)
        elif action != 'add' and config_present:
            return self.make_request('DELETE', path, data=data)

    def segment(self, name, tenant, interface_groups=[], origination='', action='add'):
        path = '/tenant[name="%s"]/segment[name="%s"]?select=name&single=true' % (tenant, name)
        response = self.make_request('GET', path, data='{}')
        config_present = True if response.status_code != 404 else False

        path = '/tenant[name="%s"]/segment[name="%s"]' %(tenant, name)
        data = '{"name": "%s", "origination": "%s"}' % (name, origination)
        if action == 'add' and not config_present:
            response = self.make_request('PUT', path, data=data)
        elif action != 'add' and config_present:
            return self.make_request('DELETE', path, data=data)
        
        if interface_groups:
            for interface_group, vlan in interface_groups:
                result = self.interface_group_segment_membership(interface_group, name, tenant, vlan=vlan)

    def get_segments(self, tenant, prefix=''):
        path = '/tenant[name="%s"]' % tenant
        response = self.make_request('GET', path, data='{}')
        segments = []
        if 'segment' in response.json()[0]:
            segments = response.json()[0]['segment']
        return [segment for segment in segments if segment['name'].startswith(prefix)]
                                                            
    def interface_group_segment_membership(self, interface_group, segment, tenant, vlan='-1', action='add'):
        """ """
        path = '/tenant[name="%s"]/segment[name="%s"]/interface-group-membership-rule[vlan=%s][interface-group="%s"]' %(tenant, segment, vlan, interface_group)
        data = '{"vlan": %s, "interface-group": "%s"}' %(vlan, interface_group)
        return self.make_request('POST' if action == 'add' else 'DELETE', path, data=data)
    
    def interface_stats(self, interface, switch_dpid, action='get'):
        """ """
        if action != 'clear':
            path = '/info/statistic/interface-counter[interface/name="%s"][switch-dpid="%s"]?select=interface[name="%s"]' % (interface, switch_dpid, interface)
            response = self.make_request('GET', path, data='{}').json()
            return response[0] if response else []
        else:
            path = '/info/statistic/interface-counter[switch-dpid="%s"]/interface[name="%s"]' % (switch_dpid, interface)
            response = self.make_request('DELETE', path, data='{}').json()
            return response

    def switch_dpid(self, switch):
        """ """
        path = '/switch-config[name="%s"]?select=dpid' % switch
        response = self.make_request('GET', path, data='{}', core_path=True)
        return response.json()[0]['dpid'].lower()

    def interface(self, switch, interface, action='no-shutdown'):
        """ """
        if action == 'shutdown':
            path = '/switch-config[name="%s"]/interface[name="%s"]' %(switch, interface)
            data = '{"shutdown": true}'
            return self.make_request('PATCH', path, data=data, core_path=True)
        elif action == 'no-shutdown':
            path = '/switch-config[name="%s"]/interface[name="%s"]/shutdown' %(switch, interface)
            return self.make_request('DELETE', path, data='{}', core_path=True)

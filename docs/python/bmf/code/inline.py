# Big Monitoring Fabric Inline config script 
import requests
import json 
import sys

requests.packages.urllib3.disable_warnings()

controller_ip = ""
username = ""
password = ""

# do not modify
cookie = ""
bigchain_path = '/api/v1/data/controller/applications/bigchain/'

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
                        
def authentication():
    global cookie
    method = 'POST'
    path = "/auth/login"
    data = '{"user":"%s", "password":"%s"}' % (username, password)
    json_content = controller_request(method, path, data)
    cookie = json_content['session_cookie']
    print 'Login successful'

def authentication_revoke():
    method = "DELETE"
    path = '/api/v1/data/controller/core/aaa/session[auth-token="%s"]' % cookie
    status_code = controller_request(method, path)

def get_controller_version():
    method = 'GET'
    path = '/rest/v1/system/version'
    data = '{}'
    json_content = controller_request(method, path, data=data, dry_run=True)
    return json_content[0] if type(json_content) == list else None
    
def add_chain(name):
    method = 'PUT'
    path = bigchain_path+'chain[name="%s"]' % name
    data = '{"name": "%s"}' % name
    controller_request(method, path, data=data, dry_run=True)

def delete_chain(name):
    method = 'DELETE'
    path = bigchain_path+'chain[name="%s"]' % name
    data = '{}'
    controller_request(method, path, data=data)

def add_chain_endpoints(chain_name, switch_alias_or_dpid, endpoint1, endpoint2):
    method = 'PATCH'
    path = bigchain_path+'chain[name="%s"]/endpoint-pair' % chain_name
    data = '{"switch": "%s", "endpoint1": "%s", "endpoint2": "%s"}' %(switch_alias_or_dpid, endpoint1, endpoint2)
    response = controller_request(method, path, data=data, dry_run=False)
    print response

def add_service(name):
    method = 'PUT'
    path = bigchain_path+'service[name="%s"]' % name
    data = '{"name": "%s"}' % name
    controller_request(method, path, data=data)

def add_service_instance(name, instance_id, switch, in_intf, out_intf):
    method = 'PUT'
    path = bigchain_path+'service[name="%s"]/instance[id=%s]/interface-pair' % (name, instance_id)
    data = '{"switch": "%s", "in": "%s", "out": "%s"}' % (switch, in_intf, out_intf)
    print controller_request(method, path, data=data, dry_run=True)

def delete_service(name):
    method = 'DELETE'
    path = bigchain_path+'service[name="%s"]' % name
    data = '{}'
    controller_request(method, path, data=data)

def get_service_policy_action(name):
    """ 
    returns service name's policy action from drop (services all traffic and drops specified traffic),
    bypass-service (services all traffic except specified traffic), or do-service (default) 
    """
    method = 'GET'
    path = bigchain_path+'service[name="%s"]/policy/action' % name
    data = '{}'
    json_content = controller_request(method, path, data=data)
    return json_content[0] if type(json_content) == list else None

def get_service_policy_type(name):
    """ returns the type of service name's policy from smtp, web-smtp, web, ssl, unknown, all, or custom """
    method = 'GET'
    path = bigchain_path+'service[name="%s"]/type' % name
    data = '{}'
    json_content = controller_request(method, path, data=data)
    return json_content[0] if type(json_content) == list else None

def service_has_custom_policy(name):
    """ returns true if service has custom policy type """
    return True if get_service_policy_type(name) == "custom" else False

def get_service_policy_ip_rules(name):
    method = 'GET'
    path = bigchain_path+'service[name="%s"]/policy/rule' % name
    data = '{}'
    json_content = controller_request(method, path, data=data)
    if json_content:
        return json_content

def get_next_sequence_number(name):
    """ returns the next available sequence number for a service rule """
    rules = get_service_policy_ip_rules(name)
    if rules:
        current_sequence = map(lambda x: x['sequence'], rules)
        next_sequence_number = 1
        while next_sequence_number in current_sequence:
            next_sequence_number += 1
        return next_sequence_number
    return 1

def verify_7_tuple_dict(ip_7_tuple_dict):
    """ """
    if len(ip_7_tuple_dict) != 7:
        print "IP 5-tuple incomplete"
        return None
    try:
        if not (ip_7_tuple_dict["ip-proto"] == 17 or ip_7_tuple_dict["ip-proto"] == 6):
            print "ip-proto must be either TCP (6) or UDP (17)"
            return None
    except KeyError:
        print "ip-proto not specified"
        return None
    return True

def find_existing_service_ip_rule(name, ip_7_tuple_dict):
    """ returns list of rules equal to rule, that are already programmed on the controller for the service  """
    rules = get_service_policy_ip_rules(name)
    if not rules:
        return []
    matching_rules = []
    for rule in rules:
        if all(item in rule.items() for item in ip_7_tuple_dict.items()):
            matching_rules.append(rule)
    return matching_rules

def add_service_ip_rule(name, ip_7_tuple_dict):
    """ 
    adds ip rule to the service 
    PUT http://127.0.0.1:8082/api/v1/data/controller/applications/bigchain/service[name="Service1"]/policy/rule[sequence=6] {"src-ip-mask": "255.255.255.255", "sequence": 6, "ether-type": 2048, "src-ip": "1.4.1.4", "dst-ip-mask": "255.255.255.255", "src-tp-port": 64402, "ip-proto": 6, "dst-ip": "1.4.6.1", "dst-tp-port": 64034}
    """
    if not verify_7_tuple_dict(ip_7_tuple_dict):
        return None
    rules = find_existing_service_ip_rule(name, ip_7_tuple_dict)
    if rules:
        return None
    sequence = get_next_sequence_number(name)
    method = 'PUT'
    path = bigchain_path+'service[name="%s"]/policy/rule[sequence=%s]' % (name, sequence)
    data = '{"sequence": %s, "ether-type": %s, "ip-proto": %s, "src-ip": "%s", "src-ip-mask": "%s", "dst-ip": "%s", "dst-ip-mask": "%s", "src-tp-port": %s, "dst-tp-port": %s}' %(sequence, 2048, ip_7_tuple_dict["ip-proto"], ip_7_tuple_dict["src-ip"], ip_7_tuple_dict["src-ip-mask"], ip_7_tuple_dict["dst-ip"], ip_7_tuple_dict["dst-ip-mask"], ip_7_tuple_dict["src-tp-port"], ip_7_tuple_dict["dst-tp-port"])
    return controller_request(method, path, data=data)

def delete_service_ip_rule(name, ip_7_tuple_dict):
    """
    deletes  ip rule from a service
    DELETE http://127.0.0.1:8082/api/v1/data/controller/applications/bigchain/service[name="Service1"]/policy/rule[sequence=3] {"src-ip-mask": "255.255.255.255", "ether-type": 2048, "src-ip": "1.4.1.4", "dst-ip-mask": "255.255.255.255", "src-tp-port": 64402, "ip-proto": 6, "dst-ip": "1.4.6.1", "dst-tp-port": 64034}
    """
    if not verify_7_tuple_dict(ip_7_tuple_dict):
        return None
    rules = find_existing_service_ip_rule(name, ip_7_tuple_dict)
    if not rules:
        return None
    method = 'DELETE'
    for rule in rules:
        path = bigchain_path+'service[name="%s"]/policy/rule[sequence=%s]' % (name, rule['sequence'])
        data = rule
        controller_request(method, path, data=data)
    
if __name__ == '__main__':
    authentication()
    # create a chain
    add_chain("Chain1")
    #delete_chain("Chain1")
    add_chain_endpoints("Chain1", "inline2", "ethernet1", "ethernet12")
    add_service("MyService")
    service_has_custom_policy("MyService")
    #delete_service("MyService")
    #get_service_policy_action("Service1")
    #print get_service_policy_ip_rules("Service1")
    #print get_next_sequence_number("Service1")
    #ip_7_tuple = {"ip-proto": 6, "src-ip": "10.10.25.36", "src-ip-mask": "255.255.255.255", "dst-ip": "56.38.123.23", "dst-ip-mask": "255.255.255.255", "src-tp-port": 42365, "dst-tp-port": 80}
    #add_service_ip_rule("Service1", ip_7_tuple)
    #delete_service_ip_rule("Service1", ip_7_tuple)
    #print get_controller_version()
    #add_service_instance('WAF', '1', '00:00:cc:37:ab:2c:97:ea', 'ethernet30', 'ethernet31')
    authentication_revoke()

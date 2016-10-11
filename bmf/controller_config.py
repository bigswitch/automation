# Big Monitoring Fabric Inline config script 
import requests
import json 
import sys

requests.packages.urllib3.disable_warnings()

controller_ip = "10.2.19.102"
username = "admin"
password = "bsn"
cookie = ""

# do not modify
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

def add_chain(name):
    method = 'PUT'
    path = bigchain_path+'chain[name="%s"]' % name
    data = '{"name": "%s"}' % name
    controller_request(method, path, data=data)

def delete_chain(name):
    method = 'DELETE'
    path = bigchain_path+'chain[name="%s"]' % name
    data = '{}'
    controller_request(method, path, data=data)

def add_chain_endpoints(chain_name, switch_alias_or_dpid, endpoint1, endpoint2):
    method = 'PATCH'
    path = bigchain_path+'chain[name="%s"]/endpoint-pair' % chain_name
    data = '{"switch": "%s", "endpoint1": "%s", "endpoint2": "%s"}' %(switch_alias_or_dpid, endpoint1, endpoint2)
    controller_request(method, path, data=data)

def add_service(name):
    method = 'PUT'
    path = bigchain_path+'service[name="%s"]' % name
    data = '{"name": "%s"}' % name
    controller_request(method, path, data=data)

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

def get_service_policy_rules(name):
    method = 'GET'
    path = bigchain_path+'service[name="%s"]/policy/rule' % name
    data = '{}'
    json_content = controller_request(method, path, data=data)
    if json_content:
        rules = []
        keys = ['sequence', 'ip-proto', 'src-ip', 'src-ip-mask', 'dst-ip', 'dst-ip-mask', 'src-tp-port', 'dst-tp-port']
        for item in json_content:
            rule = str(item.get('sequence')) + " match"
            for key in keys[1:]:
                rule += " " + str(item.get(key))
            rules.append(rule)
        return rules

if __name__ == '__main__':
    authentication()
    # create a chain
    #add_chain("Chain1")
    #delete_chain("Chain1")
    #add_chain_endpoints("Chain1", "00:00:cc:37:ab:2c:9d:68", "ethernet49", "ethernet50")
    #add_service("MyService")
    #service_has_custom_policy("BC_SSLVA_Inside")
    #delete_service("MyService")
    print get_service_policy_rules("A10-Service1")
    print get_service_policy_action("A10-Service1")
    authentication_revoke()
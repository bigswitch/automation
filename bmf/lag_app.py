# Big Monitoring Fabric Inline measure rate on LAG and
# adds member when rate exceeds threshold
import requests
import json 
import os
import time

requests.packages.urllib3.disable_warnings()

# do not modify
access_token = ''
controller_ip = '10.2.19.102'
controller_path = '/api/v1/data/controller/'
bigchain_path = '/api/v1/data/controller/applications/bigchain/'

def controller_request(method, path, data="", dry_run=False):
    if not controller_ip:
        print 'You must set controller_ip to the IP address of your controller'
    controller_url = "https://%s:8443" % controller_ip
    # append path to the controller url, e.g. "https://192.168.23.98:8443" + "/api/v1/auth/login"
    url = controller_url + path
    # if a cookie exists then use it in the header, otherwise create a header without a cookie
    if access_token:
        session_cookie = 'session_cookie=%s' % access_token
        headers = {"content-type": "application/json", 'Cookie': session_cookie}
    else:
        print "Error - no access token found"
        exit(0)

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
                        
def get_controller_version():
    method = 'GET'
    path = '/rest/v1/system/version'
    data = '{}'
    json_content = controller_request(method, path, data=data, dry_run=True)
    return json_content[0] if type(json_content) == list else None
    
def get_xmit_rate(lag_interface):
    """
    returns tx rate on lag
    """
    path = controller_path + 'core/switch/lag-interface'
    data = '{}'
    json_content = controller_request('GET', path, data=data)
    for lag in json_content:
        if lag['name'] == lag_interface:
            return lag['tx-packet-rate-60']
    print 'Did not find a lag with name %s' % lag_interface

def add_lag_interface_member(switch, lag_interface, member):
    """
    adds intf to lag
    """
    #PUT http://127.0.0.1:8080/api/v1/data/controller/core/switch-config%5Bname%3D%22inline1%22%5D/lag-interface%5Bname%3D%22vtps-lag1%22%5D/member%5Bname%3D%22ethernet37%22%5D {"name": "ethernet37"}
    path = controller_path + 'core/switch-config[name="%s"]/lag-interface[name="%s"]/member[name="%s"]' \
           % (switch, lag_interface, member)
    data = '{"name": "%s"}' % member
    return controller_request('PUT', path, data=data)

if __name__ == '__main__':
    # read access token
    access_token = os.environ['BIGSWITCHACCESSTOKEN']

    # constants for the control loop
    switch = 'inline2'
    lag1 = 'a10-lag1'
    to_add_to_lag1 = 'ethernet37'
    lag2 = 'a10-lag2'
    to_add_to_lag2 = 'ethernet38'

    threshold = 5500

    while True:
        xmit_rate = get_xmit_rate(lag1)
        if xmit_rate >= threshold:
	    #time.sleep(60)
            print 'lag xmit rate per second exceeded at %s' % xmit_rate
            result = add_lag_interface_member(switch, lag1, to_add_to_lag1)
            print 'Adding %s to %s: %s' % (to_add_to_lag1, lag1, result)
            result = add_lag_interface_member(switch, lag2, to_add_to_lag2)
            print 'Adding %s to %s: %s' % (to_add_to_lag2, lag2, result)
            break
        else:
            print 'xmit rate %s is below threshold' % xmit_rate
        time.sleep(10)
    print 'control loop exited'

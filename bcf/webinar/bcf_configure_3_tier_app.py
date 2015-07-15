import requests
import json

cookie = ""
controller_url = ""

def bcf_login(IP, username, password):
    global cookie
    global controller_url
    controller_url = "https://%s:8443" % IP
    url = controller_url + "/api/v1/auth/login"
    data = '{"user":"%s", "password":"%s"}' % (username, password)
    headers = {"content-type": "application/json"}
    r = requests.post(url, data=data, headers=headers, verify=False)
    cookie =  json.loads(r.content)['session_cookie']
    
def send_rest(action, url, data):
    global cookie
    session_cookie = 'session_cookie=%s' % cookie
    headers = {'Cookie': session_cookie, "content-type": "application/json"}
    url = controller_url + "/api/v1/data/controller/applications/bcf/" + url
    r = requests.put(url, data=data, headers=headers, verify=False)
    return r.status_code
        
def configure_tenant(name) :
    url = 'tenant[name="%s"]' % name
    data = '{"name": "%s"}' % name
    send_rest('PUT', url, data)
    
def configure_segment(name, segment, vlan, ip) :
    url = 'tenant[name="%s"]/segment[name="%s"]' % (name, segment)
    data = '{"name": "%s"}' % segment
    send_rest('PUT', url, data)
    
    url = 'tenant[name="%s"]/segment[name="%s"]/port-group-membership-rule[vlan=%s][port-group="any"]' % (name, segment, vlan)
    data = '{"vlan": %s, "port-group": "any"}' % (vlan)
    send_rest('PUT', url, data)

    url = 'tenant[name="%s"]/logical-router/segment-interface[segment="%s"]' % (name, segment)
    data = '{"segment": "%s"}' % segment
    send_rest('PUT', url, data)

    url = 'tenant[name="%s"]/logical-router/segment-interface[segment="%s"]/ip-subnet' % (name, segment)
    data = '{"ip-cidr": "%s", "private": false}' %ip
    send_rest('PUT', url, data)   
 
    print "Configured segment %s successfully" % segment


bcf_login ('10.2.18.11', 'admin', 'bsn123') 

configure_tenant ('red')
configure_segment ('red', 'web', 211, '211.0.1.1/24')
configure_segment ('red', 'app', 212, '212.0.1.1/24')
configure_segment ('red', 'db', 213, '213.0.1.1/24')


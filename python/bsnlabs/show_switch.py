i#
# show switch
#
import json
import requests
requests.packages.urllib3.disable_warnings()

# To obtain the ip address of the controller click on the controller (CNTL) icon in the topology 
# diagram and select "Device Information"

# Set the IP address of your Big Cloud Fabric controller, e.g. controller_ip = "54.198.84.132"
controller_ip = ""

controller_url = "https://" + controller_ip + ":8443"

##################################
# Login 
##################################
# First you must obtain an authentication cookie from the controller, we therefore define the login path
path = "/api/v1/auth/login"
# append the login path to the controller url to obtain the full url 
url = controller_url + path

# Define the data and headers for the HTTP request
data = '{"user": "admin", "password": "bsn123"}'
headers = {"content-type": "application/json"}

# POST request made on the Big Cloud Fabric controller
response = requests.request('POST', url, data=data, headers=headers, verify=False) 
print "Authentication response\n", response.content, "\n\n"

# Extract the cookie from the response and create a session cookie string to be used in subsequent requests
cookie = json.loads(response.content)['session_cookie']
session_cookie = 'session_cookie=%s' % cookie


##################################
# Show Switch
##################################
# We obtained this REST API call from the CLI 
# REST-SIMPLE: GET http://127.0.0.1:8080/api/v1/data/controller/applications/bcf/info/fabric/switch

# path is set to substring "/api/v1/data/controller/applications/bcf/info/fabric/switch" from the above url
path = '/api/v1/data/controller/applications/bcf/info/fabric/switch'
# we append this "show switch" path to the controller url to obtain a full url
url = controller_url + path

# There is no data to pass in for this GET request
data = ''
# the headers will contain the session cookie we obtained above via the POST request
headers = {"content-type": "application/json", 'Cookie': session_cookie}

response = requests.request('GET', url, data=data, headers=headers, verify=False)
print "show switch response\n", response.content


##################################
# Logout 
##################################
path = '/api/v1/data/controller/core/aaa/session[auth-token="'+cookie+'"]'
url = controller_url + path
headers = {"content-type": "application/json", 'Cookie': session_cookie}
# DELETE request made on the Big Cloud Fabric controller
response = requests.request('DELETE', url, headers=headers, verify=False)

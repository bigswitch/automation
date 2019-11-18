import requests
url = "https://10.8.23.15:8443/api/v1/auth/login"
data = '{"user":"admin", "password": "adminadmin"}'
headers = {"content-type": "application/json"}
r = requests.post(url, data=data, headers=headers, verify=False) 
print r.content

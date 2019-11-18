import smtplib         
import requests
import json

cookie = ""
controllerURL = "https://10.2.18.11:8443"


def bcf_login(IP, username, password):
    global cookie
    global controllerIP
    controllerIP = "https://%s:8443" % IP
    url = controllerIP + "/api/v1/auth/login"
    data = '{"user":"%s", "password":"%s"}' % (username, password)
    headers = {"content-type": "application/json"}
    r = requests.post(url, data=data, headers=headers, verify=False)
    cookie =  json.loads(r.content)['session_cookie']
    
def send_rest(action, url, data):
    global cookie
    session_cookie = 'session_cookie=%s' % cookie
    headers = {'Cookie': session_cookie, "content-type": "application/json"}
    url = controllerURL + "/api/v1/data/controller/applications/bcf" + url
    print url
    response = requests.request(action, url, data=data, headers=headers, verify=False)
    return response

def send_email(from_addr,to_addr,password,content):
    subj = "****** BCF FABRIC ERROR ALERT ******"
    message_text = "Fabric Error Seen More details Below \n %s" % content
    msg = "From: %s\nTo: %s\nSubject: %s\n\n%s" % ( from_addr, to_addr, subj, message_text )
    server = smtplib.SMTP('smtp.gmail.com:587')
    server.starttls()
    server.login(from_addr,password)
    server.sendmail(from_addr, to_addr, msg)
    server.quit()
    print "SUCCESSFUL: Notification email sent successfully"

def check_bcf_errors():
    bcf_login ('10.2.18.11', 'admin', 'bsn123')
    response = send_rest('GET', '/info/summary/fabric', {})
    response_body = json.loads(response.content)
    if response_body[0]['errors']:
        send_email ('bigswitchdemo@gmail.com', 'bigswitchdemo@gmail.com', 'bigswitch123', response.content)
        print "\n****    BCF Fabric Status: Errors Found    \n****"  
    else:    
        print "\n****    BCF Fabric Status: No Errors    ****\n"    
    print response.content    


if __name__ == '__main__':
    check_bcf_errors()

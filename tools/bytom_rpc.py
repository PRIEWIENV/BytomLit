import requests

url = "http://52.82.55.145:9888/"

def call(cmd, param):
    request_url = url + cmd
    r = requests.post(request_url, data=param)
    return r.text


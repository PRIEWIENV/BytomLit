import requests

url = "http://47.99.208.8:9888/"

def call(cmd, param):
    request_url = url + cmd
    r = requests.post(request_url, json=param)
    return r.json()


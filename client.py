import hashlib
import urllib.request
import sys


CHUNK_SIZE = 64


if sys.argv[1] == 'get':
    url = 'http://localhost:8080/chunk/' + sys.argv[2]

    try:
        with urllib.request.urlopen(url) as r:
            print('>', r.status, r.reason)
            print(r.read())
    except Exception as e:
        print('>', e.status, e.reason)

if sys.argv[1] == 'put':
    data = (sys.argv[2].encode('utf-8') + b'\x00'*64)[:64]

    h = hashlib.sha512(data).hexdigest()
    url = 'http://localhost:8080/chunk/' + h
    req = urllib.request.Request(url, method='PUT', data=data)

    try:
        with urllib.request.urlopen(req) as r:
            print('>', r.status, r.reason)
            print(url)
    except Exception as e:
        print('>', e.status, e.reason)

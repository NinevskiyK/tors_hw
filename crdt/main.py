import json
from flask import Flask, request
import sys
from server import Server

import logging
log = logging.getLogger('werkzeug')
log.setLevel(logging.ERROR)

port = sys.argv[1]
me = int(sys.argv[2])
servers = sys.argv[3:]

server = Server(servers, me)

app = Flask(__name__)

@app.route('/get_key')
def get_key():
    key = request.args.get('key')
    value = server.get_value(key)
    if value is None:
        return "not_found", 404
    return value, 200

@app.route('/set_values', methods=['PATCH'])
def set_value():
    body = request.get_json()
    body = json.loads(body)
    for key, value in body.items():
        server.set_value(key, value)

    return "ok", 200

@app.route('/patch', methods=['PATCH'])
def patch():
    server.handle_broadcast(request.data)
    return "ok", 200

app.run(host='0.0.0.0', port=port)

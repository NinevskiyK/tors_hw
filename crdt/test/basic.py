import requests
import time
import json

def send_values(replica, values):
    req = json.dumps(values)
    requests.patch(f'http://127.0.0.1:808{replica}/set_values', timeout=100, json=req, headers={'Content-Type': 'application/json'})

def get_value(replica, key):
    return requests.get(f'http://127.0.0.1:808{replica}/get_key?key={key}', timeout=100).content.decode("utf-8")

def test_basic():
    send_values(1, {"k": "v"})

    time.sleep(1)

    assert get_value(1, "k") == "v"
    assert get_value(2, "k") == "v"
    assert get_value(3, "k") == "v"

    send_values(2, {"k": "new_v"})
    send_values(1, {"new_k": "another_v"})

    time.sleep(1)

    assert get_value(1, "k") == "new_v"
    assert get_value(2, "k") == "new_v"
    assert get_value(3, "k") == "new_v"

    assert get_value(1, "new_k") == "another_v"
    assert get_value(2, "new_k") == "another_v"
    assert get_value(3, "new_k") == "another_v"


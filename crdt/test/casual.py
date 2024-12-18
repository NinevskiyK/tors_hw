import time
from basic import send_values, get_value

def test_casual():
    for _ in range(10):
        send_values(1, {"k": "1"})

        while get_value(2, "k") != "1":
            time.sleep(0.1)
        send_values(2, {"another_k": "1"})

        while get_value(1, "another_k") != "1":
            time.sleep(0.1)
        send_values(1, {"last_k": "1"})

        if get_value(3, "another_k") == "1":
            assert get_value(3, "k") == "1"
        if get_value(3, "last_k") == "1":
            assert get_value(3, "k") == "1"
            assert get_value(3, "another_k") == "1"

        send_values(1, {"k": "2", "another_k": "2", "last_k": "2"})
        time.sleep(0.1)

# tc qdisc add dev eth0 root netem delay 100ms 100ms distribution normal
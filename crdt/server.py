from typing import List
import pickle
import time
from threading import Lock, Thread
from requests import patch

from vector_clock import VectorClock, CompResult
from values import Values

class Server:
    def __init__(self, servers: List[str], me: int):
        self.clock = VectorClock(len(servers) + 1, me)
        self.values = Values()
        self.servers = servers
        self.lock: Lock = Lock()
        Thread(target=self._regular_update).start()


    def _broadcast_impl(self, server):
        with self.lock:
            msg = pickle.dumps([self.values, self.clock])
        try:
            patch(server + '/patch', msg, timeout=1)
        except:
            pass


    def _broadcast(self):
        for server in self.servers:
            Thread(target=self._broadcast_impl, args=(server,)).start()


    def _regular_update(self):
        time.sleep(2)
        while True:
            self._broadcast()
            time.sleep(2)


    def handle_broadcast(self, msg):
        msg = pickle.loads(msg)
        another_clock : VectorClock = msg[1]
        another_values : Values = msg[0]
        res = self.clock.is_bigger(another_clock)

        if another_values.values != self.values.values:
            with self.lock:
                if res == CompResult.WIN:
                    self.values.update_won(another_values)
                else:
                    print(f"Found bigger clocks: {another_clock} with vals {another_values}\n\tmy{self.clock} with {self.values}")
                    self.values.update_loose(another_values)
                self.clock.merge(another_clock)

    def get_value(self, key):
        return self.values.values.get(key, None)

    def set_value(self, key, value):
        with self.lock:
            self.clock.update()
            self.values.set_value(key, value)
        self._broadcast()

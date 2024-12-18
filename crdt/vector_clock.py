from __future__ import annotations
from enum import Enum

class CompResult(Enum):
    WIN=0
    LOOSE=1

class VectorClock:
    def __init__(self, replica_cnt, me):
        self.clock = [0 for _ in range(replica_cnt)]
        self.me = me

    def update(self):
        self.clock[self.me] += 1

    def merge(self, another: VectorClock):
        for i, (a, b) in enumerate(zip(self.clock, another.clock)):
            self.clock[i] = max(a ,b)

    def is_bigger(self, another):
        i_won = False
        a_won = False
        for a, b in zip(self.clock, another.clock):
            if a < b:
                a_won = True
            if b < a:
                i_won = True
        if not i_won ^ a_won:
            if self.me < another.me:
                return CompResult.WIN
            else:
                return CompResult.LOOSE
        if i_won:
            return CompResult.WIN
        elif a_won:
            return CompResult.LOOSE

    def __str__(self):
        return self.clock.__str__()
from __future__ import annotations

class Values:
    def __init__(self):
        self.values = {}

    def update_won(self, another: Values):
        self.values = another.values | self.values

    def update_loose(self, another: Values):
        self.values = self.values | another.values

    def set_value(self, key, value):
        self.values[key] = value

    def __str__(self):
        return self.values.__str__()
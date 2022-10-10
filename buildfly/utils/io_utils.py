#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""

"""
import os
import json


def write_to_file(filepath, line):
    if type(line) == dict:
        line = json.dumps(line)

    with open(filepath, "w") as f:
        if type(line) == str:
            f.write(line + "\n")
        elif type(line) == list:
            for l in line:
                f.write(l + "\n")
        else:
            f.write(line + "\n")


def read_file_line(filepath):
    if os.path.exists(filepath):
        line = ""
        with open(filepath, "r") as f:
            line = f.readline().strip()
        return line
    else:
        return None

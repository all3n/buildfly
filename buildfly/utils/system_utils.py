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
from collections import namedtuple
import glob


def check_command_exists(cmd):
    return os.system("which %s >/dev/null 2>&1" % cmd) == 0


def exec_cmd(cmd):
    data = None
    with os.popen(cmd) as f:
        data = f.read().strip()
    return data


def get_bfly_path(d, create = False):
    bpath = os.path.expanduser("~/.buildfly/%s" % d)
    if create and not os.path.exists(bpath):
        os.makedirs(bpath)
    return bpath



def parse_glob_files(f):
    if type(f) == str:
        return glob.glob(f, recursive=True)
    elif type(f) == list:
        out = []
        for i in f:
            out.extend(parse_glob_files(i))
        return out
    else:
        raise RuntimeError("%s glob type not support" % type(f))

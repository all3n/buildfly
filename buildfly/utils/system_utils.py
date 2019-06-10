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
import platform


def check_command_exists(cmd):
    return os.system("which %s >/dev/null 2>&1" % cmd) == 0



def exec_cmd(cmd):
    data = None
    with os.popen(cmd) as f:
        data = f.read().strip()
    return data


def get_bfly_path(d):
    return os.path.expanduser("~/.buildfly/%s" % d)




def bfly_exec(f, globals, locals):
    exec(compile(open(f, "rb").read(), f, "exec"), globals, locals )


def register_function(fun):
    globals()[fun.__name__] = fun

def sys_info():
    return platform.machine(), platform.system()


def target_ext():
    system = platform.system()
    if system == "Windows":
        return ".dll", ".lib", ".exe"
    elif system == "Darwin":
        return ".dylib", ".a", ""
    else:
        return ".so", ".a", ""
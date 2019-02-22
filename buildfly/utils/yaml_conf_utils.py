#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""

"""
import yaml
from collections import namedtuple
TargetDep = namedtuple('TargetDep', ['name', 'libdesc', 'link_type', 'libs'])


class BuildTarget(object):
    pass

class BuildConf(object):
    bins = {}
    libs = {}
    args = {}
    dependency = {}

    def __init__(self, yaml_conf):
        for k,v in yaml_conf.items():
            if type(v) == dict and 'type' in v:
                if v['type'] == 'bin':
                    self.bins[k] = self.parse_bin_or_lib(v)
                elif v['type'] == 'lib':
                    self.libs[k] = self.parse_bin_or_lib(v)
            elif k == 'dependency':
                self.dependency = v
            else:
                self.args[k] = v
    def parse_bin_or_lib(self, v):
        if 'deps' in v:
            deps = v['deps']
            if type(deps) == list:
                for d in deps:
                    dep_info = self.parse_dep(d)

    # parse dep
    def parse_dep(self, dep):
        dtype = type(d)
        if dtype == str:
            # app lib
            if dep.startswith("//"):
                pass
            else:
                dinfos = d.split(":")
        elif dtype = dict:
            pass



def yaml_conf_loader(yaml_file):
    with open(yaml_file, "r", encoding='utf-8') as yf:
        conf =  yaml.load(yf.read())
        import json; print(json.dumps(conf, indent = 2))
        return BuildConf(conf)

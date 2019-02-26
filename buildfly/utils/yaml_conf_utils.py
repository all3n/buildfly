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
from collections import namedtuple, defaultdict
TargetDep = namedtuple('TargetDep', ['name', 'libdesc', 'link_type', 'lib_name'])


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

    def __str__(self):
        return "bins:%s libs:%s args:%s dependency:%s" % (
                self.bins, self.libs, self.args, self.dependency
            )
    def parse_bin_or_lib(self, v):
        if 'deps' in v:
            deps_map = defaultdict(list)
            deps = v['deps']
            if type(deps) == list:
                for d in deps:
                    target_deps = self.parse_dep(d)
                    for tdep in target_deps:
                        deps_map[tdep.name].append(tdep)
            v['deps'] = deps_map
        return v


    # parse dep
    def parse_dep(self, dep):
        dtype = type(dep)
        target_deps = []
        if dtype == str:
            # app lib
            # //xx-lib
            if dep.startswith("//"):
                name = dep[2:]
                dep_type = "local"
                target_dep = TargetDep(name=name, libdesc = dep, link_type = "static", lib_name=name)
                target_deps.append(target_dep)
            else:
                dinfos = dep.split(":")
                len_dinfo = len(dinfos)
                # defined dep
                # "name:static:lib_name"
                if len_dinfo == 3:
                    name, link_type, lib_name = dinfos
                    target_dep = TargetDep(name=name, libdesc = name, link_type = link_type, lib_name=lib_name)
                elif len_dinfo == 2:
                    name, link_type = dinfos
                    target_dep = TargetDep(name=name, libdesc = name, link_type = link_type, lib_name=name)
                elif len_dinfo == 1:
                    name = dinfos
                    target_dep = TargetDep(name=name, libdesc = name, link_type = "static", lib_name=name)
                else:
                    raise Exception("%s,lib desc len:%d not support,must match name:link_type:lib_name" % (dep, len_dinfo))
                target_deps.append(target_dep)
        elif dtype == dict:
            for k,v in dep.items():
                typev = type(v)
                if typev == dict:
                    # "name": "static:libname"   or {'googletest': 'static:gtest'}
                    for dn,dv in v.items():
                        dv_split = dv.split(":")
                        assert len(dv_split) == 2
                        link_type, lib_name = dv_split
                        target_dep = TargetDep(name=dn, libdesc = dn, link_type = link_type, lib_name=lib_name)
                        target_deps.append(target_dep)
                elif typev == list:
                    # "name": ["static:libname"]
                    for dvs in v:
                        dv_split = dvs.split(":")
                        assert len(dv_split) == 2
                        link_type, lib_name = dv_split
                        target_dep = TargetDep(name=k, libdesc = k, link_type = link_type, lib_name=lib_name)
                        target_deps.append(target_dep)
                elif typev == str:
                    dv_split = v.split(":")
                    assert len(dv_split) == 2
                    link_type, lib_name = dv_split
                    target_dep = TargetDep(name=k, libdesc = k, link_type = link_type, lib_name=lib_name)
                    target_deps.append(target_dep)
                else:
                    raise Exception("%s,type dep only support list,dict" % (v))
        return target_deps



def yaml_conf_loader(yaml_file):
    with open(yaml_file, "r", encoding='utf-8') as yf:
        conf =  yaml.load(yf.read())
        # import json; print(json.dumps(conf, indent = 2))
        return BuildConf(conf)

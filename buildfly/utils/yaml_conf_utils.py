#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""

"""
import hashlib
import os
import sys
from collections import defaultdict

import yaml
from six.moves.urllib_parse import urlparse

from buildfly.utils.dep_utils import *
from buildfly.utils.system_utils import get_bfly_path

TargetDep = namedtuple('TargetDep', ['name', 'libdesc', 'link_type', 'lib_name'])


class BuildTarget(object):
    pass


class BuildDependency(object):
    cmds = []
    modules = []
    lib_info = None
    dep_type = None
    _cache_dir = None
    _repo_dir = None
    _install_dir = None
    _code_dir = None

    def __init__(self, name, dep_obj):
        print(dep_obj)
        self.name = name
        if type(dep_obj) == str:
            self.libdesc = dep_obj
            self.lib_info = detact_lib_type(self.libdesc)
            self.dep_type = self.lib_info['type']
            self.url = self.lib_info['url']
        elif type(dep_obj) == dict:
            if 'url' in dep_obj:
                self.dep_type = 'url'
                self.url = dep_obj['url']
            if 'cmds' in dep_obj and type(dep_obj['cmds']) == list and \
                    len(dep_obj['cmds']) > 0:
                self.cmds = dep_obj['cmds']
            if 'modules' in dep_obj:
                self.modules = dep_obj['modules']

    def save_modules(self):
        with open(self.get_install_modules_file(), "w") as mf:
            mf.write(",".join(self.modules))

    def is_modules_change(self):
        if self.modules:
            install_modules = set(self.get_install_modules())
            for m in self.modules:
                if m not in install_modules:
                    return True
            return False
        else:
            return False

    def get_install_modules(self):
        install_modules_file = self.get_install_modules_file()
        if os.path.exists(install_modules_file):
            with open(install_modules_file, "r") as mf:
                return mf.readline().strip().split(",")
        else:
            return []

    def get_install_modules_file(self):
        return os.path.join(self.get_install_dir(), "buildfly.modules")

    def get_base_dir(self, category="repo"):
        lib_info = self.lib_info
        if self.dep_type == 'github' or self.dep_type == "git":
            out_dir = get_bfly_path("{category}/{lib_type}/{owner}/{repo}/{repo_info}".format(
                category=category,
                lib_type=lib_info["type"],
                owner=lib_info['owner'],
                repo=lib_info['repo'],
                repo_info="/".join(lib_info['repo_info'])
            ))
        elif self.dep_type == 'url':
            urlparse_res = urlparse(self.url)
            url_host = urlparse_res.netloc
            url_md5 = hashlib.md5(self.url.encode('utf-8')).hexdigest()
            out_dir = get_bfly_path("{category}/{name}/{url_md5}".format(
                category=category,
                name=self.name,
                url_md5=url_md5
            ))
        else:
            print(lib_info)
            print("%s not support" % self.dep_type)
            sys.exit(-1)

        return out_dir

    def get_cache_dir(self):
        if not self._cache_dir:
            self._cache_dir = self.get_base_dir("cache")
        return self._cache_dir

    def get_repo_dir(self):
        if not self._repo_dir:
            self._repo_dir = self.get_base_dir("repo")
        return self._repo_dir

    def get_install_dir(self):
        if not self._install_dir:
            self._install_dir = self.get_base_dir("install")
        return self._install_dir

    def get_code_dir(self):
        if not self._code_dir:
            repo_dir = self.get_repo_dir()
            subdirs = os.listdir(repo_dir)
            dircnt = len(subdirs)
            if dircnt == 1:
                self._code_dir = os.path.join(repo_dir, subdirs[0])
                if not os.path.isdir(self._code_dir):
                    raise Exception("if code directory has one file,this file must be directory")
            elif dircnt == 0:
                raise Exception("code dir must not empty")
            else:
                self._code_dir = repo_dir
        return self._code_dir


class BuildConf(object):
    bins = {}
    libs = {}
    args = {}
    dependency = {}

    def __init__(self, yaml_conf):
        for k, v in yaml_conf.items():
            if type(v) == dict and 'type' in v:
                if v['type'] == 'bin':
                    self.bins[k] = self.parse_bin_or_lib(v)
                elif v['type'] == 'lib':
                    self.libs[k] = self.parse_bin_or_lib(v)
            elif k == 'dependency':
                self.parse_dependencies(v)
            else:
                self.args[k] = v

    def __str__(self):
        return "bins:%s libs:%s args:%s dependency:%s" % (
            self.bins, self.libs, self.args, self.dependency
        )

    def parse_dependencies(self, deps):
        if deps:
            for name, dep_obj in deps.items():
                self.dependency[name] = BuildDependency(name, dep_obj)

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
                target_dep = TargetDep(name=name, libdesc=dep, link_type="static", lib_name=name)
                target_deps.append(target_dep)
            else:
                dinfos = dep.split(":")
                len_dinfo = len(dinfos)
                # defined dep
                # "name:static:lib_name"
                if len_dinfo == 3:
                    name, link_type, lib_name = dinfos
                    target_dep = TargetDep(name=name, libdesc=name, link_type=link_type, lib_name=lib_name)
                elif len_dinfo == 2:
                    name, link_type = dinfos
                    target_dep = TargetDep(name=name, libdesc=name, link_type=link_type, lib_name=name)
                elif len_dinfo == 1:
                    name = dinfos
                    target_dep = TargetDep(name=name, libdesc=name, link_type="static", lib_name=name)
                else:
                    raise Exception(
                        "%s,lib desc len:%d not support,must match name:link_type:lib_name" % (dep, len_dinfo))
                target_deps.append(target_dep)
        elif dtype == dict:
            for k, v in dep.items():
                typev = type(v)
                if typev == dict:
                    # "name": "static:libname"   or {'googletest': 'static:gtest'}
                    for dn, dv in v.items():
                        dv_split = dv.split(":")
                        assert len(dv_split) == 2
                        link_type, lib_name = dv_split
                        target_dep = TargetDep(name=dn, libdesc=dn, link_type=link_type, lib_name=lib_name)
                        target_deps.append(target_dep)
                elif typev == list:
                    # "name": ["static:libname"]
                    for dvs in v:
                        dv_split = dvs.split(":")
                        assert len(dv_split) == 2
                        link_type, lib_name = dv_split
                        target_dep = TargetDep(name=k, libdesc=k, link_type=link_type, lib_name=lib_name)
                        target_deps.append(target_dep)
                elif typev == str:
                    dv_split = v.split(":")
                    assert len(dv_split) == 2
                    link_type, lib_name = dv_split
                    target_dep = TargetDep(name=k, libdesc=k, link_type=link_type, lib_name=lib_name)
                    target_deps.append(target_dep)
                else:
                    raise Exception("%s,type dep only support list,dict" % (v))
        return target_deps


def yaml_conf_loader(yaml_file):
    with open(yaml_file, "r", encoding='utf-8') as yf:
        try:
            conf = yaml.load(yf.read())
        except Exception as e:
            print(e)
            print("yaml load fail")
            sys.exit(-1)
        # import json; print(json.dumps(conf, indent = 2))
        return BuildConf(conf)

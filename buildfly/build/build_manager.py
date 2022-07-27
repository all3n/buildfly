#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghch <wanghch@wanghch-MacBookPro.local>
#
# Distributed under terms of the MIT license.

"""

"""
import importlib
import glob
import os
from buildfly.utils.string_utils import camelize
from buildfly.common import BFlyManifest
from buildfly.env import BENV
from buildfly.config.pkg_config import PkgConfigFile


class BuildManager(object):
    def __init__(self):
        pass

    def build(self, app_dep):
        code_dir = app_dep.get_code_dir()
        install_dir = app_dep.get_install_dir()
        cmds = app_dep.cmds
        if cmds:
            build_type = 'custom'
        else:
            build_type = self.detact_build_type(code_dir)
        build_class = build_type + "_build"
        build_module = importlib.import_module("buildfly.build." + build_class)
        build_obj = getattr(build_module, camelize(build_class))()
        build_obj.build(app_dep, code_dir, install_dir)

    def detact_build_type(self, code_dir):
        code_files = os.listdir(code_dir)
        if "CMakeLists.txt" in code_files:
            return "cmake"
        elif "configure" in code_files:
            return "configure_make"
        elif "makefile" in code_files or "Makefile" in code_files:
            return "makefile"
        else:
            raise Exception("unsupport build type")

    def write_manifest(self, app_dep):
        install_dir = app_dep.get_install_dir()

        libs_dirs = ['lib', 'lib64']
        pcf = None
        for ld in libs_dirs:
            pkdir = os.path.join(install_dir, ld, 'pkgconfig')
            if os.path.exists(pkdir):
                pcs = glob.glob(os.path.join(pkdir, '*.pc'))
                for pc in pcs:
                    pcf = PkgConfigFile(pc)
        meta = {
            "name": app_dep.name,
            "repo": {
                "url": app_dep.url,
                "type": app_dep.dep_type
            },
            "prefix": install_dir,
            "system": BENV.system,
            "arch": BENV.machine,
            "libc_version": BENV.libc_ver
        }
        if pcf:
            fields = ['includedir', 'libdir', 'version', 'description', 'libs', 'cflags', 'exec_prefix:bin_path']
            for f in fields:
                name, alias = f.split(":") if ':' in f else (f, f)
                if pcf.has(f):
                    meta[alias] = pcf.get(name)

        with open(os.path.join(install_dir, "manifest.json"), "w") as f:
            bm = BFlyManifest.from_dict(meta)
            f.write(bm.to_json())

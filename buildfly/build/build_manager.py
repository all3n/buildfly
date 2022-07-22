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
import os


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
        build_obj = getattr(build_module, build_class)()
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

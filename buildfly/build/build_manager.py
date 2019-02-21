#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghch <wanghch@wanghch-MacBookPro.local>
#
# Distributed under terms of the MIT license.

"""

"""
import os
import importlib

class build_manager(object):
    def __init__(self):
        pass

    def build(self, code_dir, install_dir):
        build_type = self.detact_build_type(code_dir)
        build_class = build_type + "_build"
        build_module = importlib.import_module("buildfly.build." + build_class)
        build_obj = getattr(build_module, build_class)
        build_obj.build(code_dir, install_dir)



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





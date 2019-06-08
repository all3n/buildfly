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
from buildfly.utils.labels_utils import parse_label
from buildfly.generator.makefile_generator import makefile_generator
class build_manager(object):
    package = None
    targets = {}
    vars = {}
    flags = {}
    build_dir = None
    generator = "makefile"
    def __init__(self):
        pass

    def build(self,app_dep):
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



    def register_target(self, target):
        abs_label = parse_label(target.name, self.package)
        if target.deps:
            target.deps = [parse_label(l, self.package) for l in target.deps]
        print("register %s  =>  %s" % (abs_label, target))
        self.targets[abs_label] = target

    def register_var(self, name, var):
        self.vars[name] = var

    def register_func(self , fun):
        self.register_var(fun.__name__, fun)

    def get_exports(self):
        return self.vars

    def check_deps_valid(self):
        for label, target in self.targets.items():
            if not target.deps:
                continue
            for dep in target.deps:
                if dep == label:
                    raise RecursionError("circular dependencies %s" % dep)
                if dep not in self.targets:
                    raise LookupError("%s dep not exist" % (dep))

    def build_target(self, label):
        label = parse_label(label, "//")
        if label not in self.targets:
            raise LookupError("invalid target label")
        target = self.targets[label]
        if target.deps:
            # build deps first
            for dep in target.deps:
                if dep in self.flags:
                    continue
                self.build_target(dep)
        # build target

        print("start build target %s" % label)
        self.generator.write(target)

        #...
        self.flags[label] = True


    def run(self, label = None):
        if self.generator == "makefile":
            gen_file = os.path.join(self.build_dir, "Makefile")
            self.generator = makefile_generator(gen_file)
        if label:
            self.build_target(label)
        else:
            # build all target
            for l, target in self.targets.items():
                self.build_target(l)


        if self.generator:
            self.generator.close()







BUILD_MANAGER = build_manager()

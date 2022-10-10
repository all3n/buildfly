#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""

"""
import glob
import os
import re
import stat
import sys

from buildfly.actions.base_action import BaseAction
from buildfly.utils.color_utils import *
from buildfly.utils.dep_utils import get_dep, get_glibc, get_dep_compile_options, check_if_needed
from buildfly.utils.system_utils import *
from buildfly.backend import *
from buildfly.utils.yaml_conf_utils import yaml_conf_loader, BuildDependency
from buildfly.utils.api_utils import BuildFlyAPI, bfly_api_method
from buildfly.common import BFlyRepo, BFlyBin, BFlyLibrary, BFlyDep
from buildfly.utils.string_utils import camelize

CONF_NAME = "buildfly.yaml"
CONF_SCRIPT = "bfly_workspace.py"
COMPILER_PATTERN = re.compile("(\w+)([<=>]{1,2})?([\d\.\w]+)?")
LIBC_VERSION_PATTERN = re.compile("libc-([\d\.]+)\.so")
import logging

logger = logging.getLogger(__name__)


class BuildAction(BaseAction):

    def __init__(self) -> None:
        self.backend_instance = None
        self.build_dir = self.get_cur_file('build')
        self.mode = "Debug"
        self.toolchain = None
        self.bins = {}
        self.libs = {}
        self.deps = {}
        self.repos = {}
        self.callbacks = {}
        # cmake/makefile/ninja
        self.backend = "cmake"
        self.on_after_build = None
        self.on_before_build = None


    def parse_args(self, parser):
        parser.add_argument('target', metavar='target', type=str, nargs="?",
                            help="build target")

    def run(self, code_dir=None):
        cur_dir = os.path.abspath(sys.path[0]) if code_dir is None else code_dir
        buildfly_script = os.path.join(cur_dir, CONF_SCRIPT)
        if os.path.exists(buildfly_script):
            with BuildFlyAPI(self):
                with open(buildfly_script, "rb") as f:
                    exec("from buildfly.api import *\n" + f.read().decode("utf-8"))
                    backend_cls = eval(camelize(self.backend) + "Backend")
                    self.backend_instance = backend_cls(self)
                    self.backend_instance.setup()
                    self.install_deps()
                    if self.on_before_build:
                        logger.info("before build")
                        self.on_before_build()
                    logger.info("start build")
                    self.backend_instance.generate()
                    self.backend_instance.build()
                    # self.start_build()
                    if self.on_after_build:
                        logger.info("after build")
                        self.on_after_build()
        else:
            logger.error(f"{CONF_SCRIPT} not found")


    @bfly_api_method
    def set(self, name, value):
        setattr(self, name, value)

    @bfly_api_method
    def get(self, name, def_val=None):
        return getattr(self, name, def_val)

    @bfly_api_method
    def set_backend(self, backend):
        self.backend = backend

    @bfly_api_method
    def set_mode(self, mode):
        self.mode = mode

    @bfly_api_method
    def set_toolchain(self, toolchain):
        self.toolchain = toolchain

    @bfly_api_method
    def add_binary(self, name, **kwargs):
        self.bins[name] = BFlyBin(**kwargs)

    @bfly_api_method
    def add_library(self, name, **kwargs):
        self.libs[name] = BFlyLibrary(**kwargs)

    @bfly_api_method
    def set_build_dir(self, build_dir):
        self.build_dir = build_dir

    @bfly_api_method
    def add_dep(self, name, artifact_id=None, **kwargs):
        kwargs.update({"name": name, "artifact_id": artifact_id})
        self.deps[name] = BFlyDep(**kwargs)

    @bfly_api_method
    def config_repo(self, dep, **kwargs):
        self.repos[dep] = dict(kwargs)

    @bfly_api_method
    def before_build(self, fn):
        setattr(self, "on_before_build", fn)

    @bfly_api_method
    def after_build(self, fn):
        setattr(self, "on_after_build", fn)

    def install_deps(self):
        for name, dep in self.deps.items():
            x = check_if_needed(dep)
            if x:
                name, dep = x
                bd = BuildDependency(name, dep_obj=dep)
                get_dep(bd)

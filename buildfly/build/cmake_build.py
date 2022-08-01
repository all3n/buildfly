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

from buildfly.build.basic_build import BasicBuild
from buildfly.env import BENV
from buildfly.config.global_config import G_CONFIG

class CmakeBuild(BasicBuild):
    build_dir = "build-buildfly"
    nproc = BENV.cpu_count
    def build(self, app_dep, code_dir, install_dir_path):
        print("cmake %s" % code_dir)

        CMD = f"cd %s;mkdir {self.build_dir}; cd {self.build_dir}; {self.gcc_env} cmake -D CMAKE_INSTALL_PREFIX=%s ..; make -j{self.nproc}; make install" % (
        code_dir, install_dir_path)
        print(CMD)
        os.system(CMD)

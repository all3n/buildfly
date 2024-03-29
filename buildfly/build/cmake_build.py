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


class CmakeBuild(BasicBuild):
    def build(self, app_dep, code_dir, install_dir_path):
        print("cmake %s" % code_dir)
        CMD = "cd %s;mkdir build; cd build; cmake -D CMAKE_INSTALL_PREFIX=%s ..; make; make install" % (
        code_dir, install_dir_path)
        print(CMD)
        os.system(CMD)

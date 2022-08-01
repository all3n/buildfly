#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghch <wanghch@wanghch-MacBookPro.local>
#
# Distributed under terms of the MIT license.

"""

"""
from buildfly.config.global_config import G_CONFIG


class BasicBuild(object):
    def __init__(self):
        gcc_home = G_CONFIG.get_value("gcc.home")
        self.gcc_env = f"CC={gcc_home}/bin/gcc CXX={gcc_home}/bin/g++"
    def build(self, app_dep, code_dir, install_dir_path):
        pass

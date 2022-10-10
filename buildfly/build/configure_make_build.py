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


class ConfigureMakeBuild(BasicBuild):
    def build(self, bpkg, code_dir, install_dir_path, build_mode):
        print("configure_make %s" % code_dir)
        CMD = f"cd %s;{self.gcc_env} ./configure --prefix=%s; make; make install" % (code_dir, install_dir_path)
        print(CMD)
        os.system(CMD)

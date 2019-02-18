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
from buildfly.build.basic_build import basic_build

class configure_make_build(basic_build):
    def build(code_dir, install_dir_path):
        print("configure_make %s" % code_dir)
        CMD="cd %s;./configure --prefix=%s; make; make install" % (code_dir, install_dir_path)
        print(CMD)
        os.system(CMD)


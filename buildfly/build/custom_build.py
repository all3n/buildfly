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


class CustomBuild(BasicBuild):
    def build(self, app_dep, code_dir, install_dir_path):
        build_script = self.write_custom_build_script(code_dir, install_dir_path, app_dep)
        CMD = "cd %s;bash %s;" % (code_dir, build_script)
        print(CMD)
        os.system(CMD)

    def write_custom_build_script(self, code_dir, install_dir_path, app_dep):
        cmds = app_dep.cmds
        modules = app_dep.modules

        build_shell_file = os.path.join(code_dir, "build_fly.build.sh")
        # write env
        with open(build_shell_file, "w") as f:
            f.write("set -x\n")
            f.write("CODE_DIR=%s\n" % code_dir)
            f.write("INSTALL_PREFIX=%s\n" % install_dir_path)
            if modules:
                f.write("INSTALL_MODULES=%s\n" % ",".join(modules))
            for cmd in cmds:
                f.write(cmd + "\n")
        return build_shell_file

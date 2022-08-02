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

import semver

from buildfly.build.basic_build import BasicBuild
from buildfly.env import BENV
from buildfly.config.global_config import G_CONFIG
import glob
import logging

logger = logging.getLogger(__name__)


class CmakeBuild(BasicBuild):
    build_dir = "build-buildfly"
    nproc = BENV.cpu_count
    libc_17 = semver.VersionInfo.parse("2.17.0")

    def check_rt_header(self, code_dir):
        hrs = glob.glob(f"{code_dir}/**/*.h")
        use_ctime = False
        for h in hrs:
            ah = open(h, 'r').read()
            if "<ctime>" in ah or "<time.h>" in ah:
                use_ctime = True
        return use_ctime

    def gen_build_script(self, code_dir, install_dir_path):
        build_script = os.path.join(code_dir, "bfly_build_script.sh")

        with open(build_script, "w") as bf:
            if self.gcc_home:
                bf.write(f"export CC={self.gcc_home}/bin/gcc\n")
                bf.write(f"export CXX={self.gcc_home}/bin/g++\n")
            if BENV.is_linux() and BENV.libc_version < self.libc_17:
                if self.check_rt_header(code_dir):
                    logger.info(f"libc {BENV.libc_version} < 2.17,add -lrt to CFLAGS,CXXFLAGS")
                    bf.write("export CFLAGS=-lrt\n")
                    bf.write("export CXXFLAGS=-lrt\n")

            bf.write(f"rm -rf {self.build_dir}\n")
            bf.write(f"mkdir -p {self.build_dir}\n")
            bf.write(f"cd {self.build_dir}\n")
            bf.write(f"cmake -D CMAKE_INSTALL_PREFIX={install_dir_path} ..\n")
            bf.write(f"make -j{self.nproc}\n")
            bf.write(f"make install\n")
        return build_script

    def build(self, app_dep, code_dir, install_dir_path):
        print("cmake %s" % code_dir)
        build_script = self.gen_build_script(code_dir, install_dir_path)
        CMD = f"cd {code_dir};bash {build_script}"
        os.system(CMD)

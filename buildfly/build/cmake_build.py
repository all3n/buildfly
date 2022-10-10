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

from buildfly.utils.log_utils import get_logger

logger = get_logger(__name__)


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

    def gen_build_script(self, code_dir, install_dir_path, build_mode):
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

            if 'envs' in self.params:
                envs = self.params['envs']
                for name, v in envs.items():
                    bf.write(f"export {name}={v}\n")
            cmake_parmas = []
            if 'cmake' in self.params:
                pcmake = self.params['cmake']
                if 'variable' in pcmake:
                    cmake_vars = pcmake['variable']
                    for name, v in cmake_vars.items():
                        cmake_parmas.append(f"-D{name}={v}")

            bf.write(f"rm -rf {self.build_dir}\n")
            bf.write(f"mkdir -p {self.build_dir}\n")
            bf.write(f"cd {self.build_dir}\n")

            cmake_parmas_str = " ".join(cmake_parmas)

            cmake_cmd = f"cmake -D CMAKE_INSTALL_PREFIX={install_dir_path} -D CMAKE_BUILD_TYPE={build_mode} {cmake_parmas_str} .."
            logger.info(f"{cmake_cmd}")
            bf.write(f"{cmake_cmd}\n")
            bf.write(f"make clean\n")
            bf.write(f"make -j{self.nproc}\n")
            bf.write(f"make install\n")
        return build_script

    def build(self, bpkg, code_dir, install_dir_path, build_mode):
        print("cmake %s" % code_dir)
        build_script = self.gen_build_script(code_dir, install_dir_path, build_mode)
        CMD = f"cd {code_dir};bash {build_script}"
        os.system(CMD)

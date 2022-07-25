#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2022 wanghch <wanghch@wanghch-MacBookPro.local>
#
# Distributed under terms of the MIT license.

"""

"""
import logging
import platform
from buildfly.utils.system_utils import exec_cmd

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)


class BuildFlyEnv(object):
    def __init__(self):
        # ('glibc', '2.29')
        self.libc_ver = platform.libc_ver()
        # x86_64
        self.machine = platform.machine()
        # only work in mac
        self.mac_ver = platform.mac_ver()
        # Linux-5.4.0-122-generic-x86_64-with-glibc2.29
        self.platform = platform.platform()
        #
        self.cpu_count = platform.os.cpu_count()
        # Linux
        self.system = platform.system().lower()

    def is_windows(self):
        return self.system == "windows"

    def is_linux(self):
        return self.system == "linux"

    def is_macos(self):
        return self.system == "Darwin"

    def detect_compiler(self):
        if self.is_macos():
            clang_v = exec_cmd("clang -v 2>&1|grep version")
            if clang_v:
                self.clang_version = clang_v.split(" ")[3]
            else:
                self.clang_version = None
            logger.info("clang version %s" % self.clang_version)
        elif self.is_linux():
            gcc_v = exec_cmd("gcc -v 2>&1|grep 'gcc version'")
            if gcc_v:
                self.gcc_version = gcc_v.split(" ")[2]
            else:
                self.gcc_version = None
            logger.info("gcc version %s" % self.gcc_version)

    def detect_cmake(self):
        logger.info("detect cmake")
        if self.is_linux() or self.is_macos():
            cmake_v = exec_cmd("cmake --version 2>&1|grep 'cmake version'")
            if cmake_v:
                self.cmake_version = cmake_v.split(" ")[2]
            else:
                self.cmake_version = None
            logger.info("cmake version %s" % self.cmake_version)

    def detect(self):
        self.detect_compiler()
        self.detect_cmake()


BENV = BuildFlyEnv()
BENV.detect()

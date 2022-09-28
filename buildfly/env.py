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
import re
import sys

import semver

from buildfly.utils.system_utils import exec_cmd

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)
LIBC_VERSION_PATTERN = re.compile("libc-([\d\.]+)\.so")


class BuildFlyEnv(object):
    def __init__(self):
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
        self.detect_flag = False

    def is_windows(self):
        return self.system == "windows"

    def is_linux(self):
        return self.system == "linux"

    def is_macos(self):
        return self.system == "darwin"

    # not found attr
    def __getattr__(self, item):
        if item == "cmake_version":
            self.detect_cmake()
        elif item == "libc_version":
            self.check_glibc()
        elif item == "gcc_version" or item == "clang_version":
            self.detect_compiler()
        return super(BuildFlyEnv, self).__getattribute__(item)

    def check_glibc(self):
        self.libc_version = None
        if self.is_linux():
            libc_so_file = exec_cmd("readlink -f `ldconfig -p|grep libc.so.6|head -n 1|awk -F\"=> \" '{print $2}'`")
            version_match = LIBC_VERSION_PATTERN.search(libc_so_file)
            if version_match:
                libc_version = version_match.group(1)
                libc_vers = libc_version.split(".")
                if len(libc_vers) == 2:
                    libc_vers.append(0)
                self.libc_version = semver.VersionInfo(*libc_vers)
                logger.info("libc version:%s" % self.libc_version)

    def detect_compiler(self):
        self.gcc_version = None
        self.clang_version = None
        if self.is_macos():
            clang_v = exec_cmd("clang -v 2>&1|grep version")
            if clang_v:
                self.clang_version = clang_v.split(" ")[3]
            logger.info("clang version %s" % self.clang_version)
        elif self.is_linux():
            gcc_v = exec_cmd("gcc -v 2>&1|grep 'gcc version'")
            if gcc_v:
                self.gcc_version = gcc_v.split(" ")[2]
            logger.info("gcc version %s" % self.gcc_version)

    def detect_cmake(self):
        logger.info("detect cmake")
        self.cmake_version = None
        if self.is_linux() or self.is_macos():
            cmake_v = exec_cmd("cmake --version 2>&1|grep 'cmake version'")
            if cmake_v:
                self.cmake_version = cmake_v.split(" ")[2]
            logger.info("cmake version %s" % self.cmake_version)

    def detect(self):
        if not self.detect_flag:
            self.check_glibc()
            self.detect_compiler()
            self.detect_cmake()
            self.detect_flag = True


BENV = BuildFlyEnv()
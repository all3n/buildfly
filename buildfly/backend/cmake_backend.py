import os.path
import sys

from buildfly.backend.base_backend import BaseBackend
from buildfly.env import BENV
from buildfly.utils.github_api_utils import api_client
from buildfly.utils.system_utils import get_bfly_path, exec_cmd
from buildfly.utils.http_pkg_utils import download_http_pkg
import semver
import logging
import tempfile

logger = logging.getLogger(__name__)


class CmakeBackend(BaseBackend):
    def __init__(self, ctx):
        super().__init__(ctx)
        self.name = "cmake"
        self.cmake_expression = self.ctx.get('cmake_version')
        self.cmake_bin = "cmake"
        self.cmake_generator = self.ctx.get('cmake_generator', 'make').lower()

    def setup(self):
        self.cmake_bin = self.install_tool_if_required("cmake", "Kitware/CMake", self.cmake_expression, "cmake",
                                                       getattr(BENV, "cmake_version"),
                                                       pkg_os_pattern={
                                                           "linux": "linux-x86_64.tar.gz",
                                                           "darwin": "macos-universal.tar.gz"
                                                       }, bin_path={
                "linux": "bin/cmake",
                "darwin": "CMake.app/Contents/bin/cmake"
            })
        if self.cmake_generator == "ninja":
            self.ninja_bin = self.install_tool_if_required("ninja", "ninja-build/ninja",
                                                           self.ctx.get("ninja_version", "latest"),
                                                           "ninja",
                                                           None,
                                                           pkg_os_pattern={
                                                               "linux": "linux.zip",
                                                               "darwin": "mac.zip"
                                                           }, bin_path={
                    "linux": "ninja",
                    "darwin": "ninja"
                })



    def generate(self):
        cur_dir = os.path.abspath(sys.path[0])
        cur_dir_cmakefile = os.path.join(cur_dir, 'CMakeLists.txt')
        print(self.ctx.bins)
        with open(cur_dir_cmakefile, "w") as f:
            f.write("cmake_minimum_required (VERSION 3.8)\n")
            f.write("project(app)\n")
            f.write("set(APP_NAME test)\n")
            f.write("set(CMAKE_CXX_STANDARD 11)\n")
            for name, bin in self.ctx.bins.items():
                f.write("set(%s_SRCS %s)\n" % (name, " ".join(bin.get_all_srcs())))
                f.write("set(%s_INCLUDES %s)\n" % (name, " ".join(bin.includes)))
                f.write("add_executable(%s ${%s_SRCS})\n" % (name, name))
                f.write("target_include_directories(%s PUBLIC ${%s_INCLUDES})" % (name, name))

    def build(self):
        cur_dir = os.path.abspath(sys.path[0])
        build_dir = os.path.join(cur_dir, "build")
        if not os.path.exists(build_dir):
            os.makedirs(build_dir)
        if self.cmake_generator == "make":
            cmd = f"cd {build_dir}; %s ..;make -j$(nproc)" % (self.cmake_bin)
        elif self.cmake_generator == "ninja":
            cmd = f"cd {build_dir}; PATH=%s:$PATH %s -GNinja ..;%s -j$(nproc)" % (os.path.dirname(self.ninja_bin), self.cmake_bin, self.ninja_bin)
            print(cmd)
        os.system(cmd)


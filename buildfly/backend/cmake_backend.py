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
        self.cmake_bin = "cmake"

    def setup(self):
        self.cmake_expression = self.ctx.get('cmake_version')
        cmake_dir = get_bfly_path("tools/cmake")
        match = False
        match_cmake = None
        if os.path.exists(cmake_dir):
            cmake_vers = os.listdir(cmake_dir)
            cvs = sorted(list(map(semver.VersionInfo.parse, cmake_vers)), reverse=True)

            for cv in cvs:
                if cv.match(self.cmake_expression):
                    match = True
                    match_cmake = os.path.join(cmake_dir, str(cv), "bin", 'cmake')
                    break

        else:
            sv = semver.VersionInfo.parse(BENV.cmake_version)
            match = sv.match(self.cmake_expression)
            if match:
                match_cmake = 'cmake'
            logger.info(f"system cmake version {sv} [X]")
        if match:
            logger.info(f'match cmake: {match_cmake}')
            self.cmake_bin = match_cmake
        if not match:
            self.install_cmake()

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
            f.write("add_executable(%s ${%s_SRCS})\n" % (name, name))

    def build(self):
        cur_dir = os.path.abspath(sys.path[0])
        build_dir = os.path.join(cur_dir, "build")
        if not os.path.exists(build_dir):
            os.makedirs(build_dir)
        cmd = f"cd {build_dir}; cmake ..;make -j$(nproc)"
        os.system(cmd)

    def install_cmake(self):
        releases = api_client.list_releases("Kitware", "CMake")
        cmake_version = None
        cmake_assets = None
        for rv, assets in releases:
            sv = semver.VersionInfo.parse(rv.replace("v", ""))
            if sv.match(self.cmake_expression):
                cmake_version = rv
                cmake_assets = assets
                break
        logger.info(f"try install {cmake_version}")
        if BENV.is_linux():
            os_pattern = "linux-x86_64.tar.gz"
        elif BENV.is_macos():
            os_pattern = "macos-universal.tar.gz"
        asset_urls = [ast["browser_download_url"] for ast in cmake_assets if os_pattern in ast["name"]]
        logger.info(f"{asset_urls}")
        if asset_urls:
            cmake_version_dir = get_bfly_path("tools/cmake/%s/" % str(cmake_version.replace("v", "")))
            if not os.path.exists(cmake_version_dir):
                os.makedirs(cmake_version_dir)

            tmp_file = tempfile.NamedTemporaryFile(prefix=cmake_version, suffix=".tar.gz")
            # use first
            # TODO
            asset_url = asset_urls[0]
            logger.info("Download Release %s " % asset_url)
            tmp_file_path = tmp_file.name
            download_http_pkg(asset_url, tmp_file_path)
            cmd = f"tar --strip-components=1 -zxvf {tmp_file_path} -C {cmake_version_dir}"
            exec_cmd(cmd)

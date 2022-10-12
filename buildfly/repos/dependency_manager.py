#!/usr/bin/env python
# -*- coding: utf-8 -*-
# Author wanghuacheng
# Description
# Date 10/8/22 5:52 PM
# Version 1.0
import enum
import importlib
import json
import os.path

from typing import Dict

from buildfly.actions.build_action import BuildAction
from buildfly.build.build_manager import BuildManager
from buildfly.env import BENV
from buildfly.utils.compress_utils import uncompress_tar_gz
from buildfly.utils.http_pkg_utils import download_http_pkg
from buildfly.utils.io_utils import write_to_file
from buildfly.utils.log_utils import get_logger
from buildfly.repos.repo_cache_db import repo_cache
from buildfly.utils.github_api_utils import api_client
import hashlib

from collections import namedtuple

from buildfly.utils.string_utils import camelize
from buildfly.utils.system_utils import get_bfly_path
from buildfly.repos.common import BFlyPkg

logger = get_logger(__name__)


class DependencyPlugin(object):
    def __init__(self):
        self.build_manager = BuildManager()

    def tags_info(self, bpkg: BFlyPkg):
        return []

    def list_versions(self, bpkg: BFlyPkg):
        return []

    def download_pkg(self, bpkg: BFlyPkg):
        return

    def detact_build_type(self, code_dir):
        code_files = os.listdir(code_dir)
        if "CMakeLists.txt" in code_files:
            return "cmake"
        elif "configure" in code_files:
            return "configure_make"
        elif "makefile" in code_files or "Makefile" in code_files:
            return "makefile"
        else:
            raise Exception("unsupport build type")

    def build_pkg(self, build_action: BuildAction, bpkg: BFlyPkg, build_force=False, build_mode="Release"):
        # logger.info(f"{bpkg}")
        commit_sha = bpkg.ver_info['commit_sha']
        code_dir = get_bfly_path(f"repo/{bpkg.path()}")
        params_md5 = hashlib.md5(json.dumps(bpkg.meta['params']).encode("utf-8")).hexdigest()
        bpkg.param_hash = params_md5
        artifact_path = f"{bpkg.path()}/{params_md5}/{build_mode}"
        install_dir = get_bfly_path(f"install/{artifact_path}")
        pkg_metadata = os.path.join(install_dir, 'metadata')

        if not os.path.exists(code_dir):
            logger.error(f"{code_dir} not exists")
            return
        bpkg.artifact_path = artifact_path
        if build_force or not os.path.exists(pkg_metadata):
            build_type = self.detact_build_type(code_dir)
            build_class = build_type + "_build"
            build_module = importlib.import_module("buildfly.build." + build_class)
            build_obj = getattr(build_module, camelize(build_class))(bpkg.meta['params'])
            build_obj.build(bpkg, code_dir, install_dir, build_mode)
            write_to_file(pkg_metadata, bpkg.__dict__)
            repo_cache.add_pkg(bpkg)
        else:
            logger.info(f"{artifact_path}/metadata exist,skip")
        return bpkg


class GithubPlugin(DependencyPlugin):
    def __init__(self, api_client):
        self.api_client = api_client

    def tags_info(self, bpkg: BFlyPkg):
        group = bpkg.group
        if not group:
            group, _ = bpkg.meta['url'].split("/")[-2:]
            bpkg.group = group
        if bpkg.version:
            all_tags = self.api_client.list_tags(group, bpkg.name)
            ver_info = all_tags[bpkg.version]
            ver = {
                'name': ver_info['name'],
                'tarball_url': ver_info['tarball_url'],
                'commit_sha': ver_info['commit']['sha']
            }
            return ver
        else:
            kwargs = bpkg.kwargs
            branch = kwargs.get('branch', 'master')
            branch_info = self.api_client.get_branch_info(group, bpkg.name, branch)
            commit_sha = branch_info['commit']['sha']
            bpkg.commit_sha = commit_sha
            commit_tar_url = f'https://github.com/{group}/{bpkg.name}/archive/{commit_sha}.tar.gz'
            # branch_tar_url = f'https://github.com/{group}/{bpkg.name}/archive/refs/heads/{branch}.tar.gz'
            ver = {
                'name': branch,
                'tarball_url': commit_tar_url,
                'commit_sha': commit_sha
            }
            return ver

    def list_versions(self, bpkg: BFlyPkg):
        ui = bpkg.meta['url'].split("/")
        group, name = ui[-2], ui[-1]
        tags = list(api_client.list_tags(group, name).keys())
        return tags

    def download_pkg(self, bpkg: BFlyPkg):
        if bpkg.version is None and bpkg.commit_sha is None:
            bpkg.ver_info = self.tags_info(bpkg)
            bpkg.commit_sha = bpkg.ver_info['commit_sha']
        commit_sha = bpkg.commit_sha
        cache_dir = get_bfly_path(f"cache/{bpkg.path()}", True)
        code_dir = get_bfly_path(f"repo/{bpkg.path()}", True)
        # download
        tmp_pkg_file = f"{cache_dir}/code.tar.gz"
        sha_file = f"{cache_dir}/COMMIT_SHA"
        if os.path.exists(sha_file):
            logger.info(f"{bpkg.path()} tar exist,skip download")
        else:
            if not hasattr(bpkg, 'ver_info'):
                bpkg.ver_info = self.tags_info(bpkg)
            tarball_url = bpkg.ver_info['tarball_url']
            if download_http_pkg(tarball_url, tmp_pkg_file):
                write_to_file(sha_file, commit_sha)

        # uncompress
        code_sha = f"{code_dir}/COMMIT_SHA"
        if os.path.exists(code_sha):
            logger.info(f"{bpkg.path()} code exist,skip uncompress")
        else:
            uncompress_tar_gz(code_dir, tmp_pkg_file, 1)
            write_to_file(code_sha, commit_sha)


class DependencyMananger(object):

    def __init__(self):
        self.plugins: Dict[str, DependencyPlugin] = {}
        self.build_action = BuildAction()

    def register(self, name, plugin):
        self.plugins[name] = plugin

    def detect_type(self, bpkg: BFlyPkg):
        purl = bpkg.meta['url']
        if 'github.com' in purl:
            return 'github'
        elif purl.startswith("git@"):
            return "git"
        elif purl.startswith("http"):
            return "url"

        # def parse_pkg(self, name, group, version):

    def parse_pkg(self, bpkg):
        # logger.info("%s %s %s", name, group, version)
        name = bpkg.name
        bpkg.meta = repo_cache.get_pkg(name)
        bpkg.type = self.detect_type(bpkg)
        assert bpkg.type in self.plugins, f"plugin {bpkg.type} not support"
        bpkg.arch = BENV.machine
        bpkg.os = BENV.system
        bpkg.libc_version = str(BENV.libc_version)
        bpkg.libstdcxx_version = BENV.libstdcxx_version

        return bpkg

    def list_versions(self, bpkg: BFlyPkg):
        return self.plugins[bpkg.type].list_versions(bpkg)

    def download_pkg(self, bpkg: BFlyPkg):
        return self.plugins[bpkg.type].download_pkg(bpkg)

    def build_pkg(self, bpkg: BFlyPkg, build_force=False, build_mode="Release"):
        return self.plugins[bpkg.type].build_pkg(self.build_action, bpkg, build_force, build_mode)
        pass

    def install_pkg(self):
        pass


DEP_MGR = DependencyMananger()
DEP_MGR.register("github", GithubPlugin(api_client))

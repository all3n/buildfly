#!/usr/bin/env python
# -*- coding: utf-8 -*-
# Author wanghuacheng
# Description
# Date 10/8/22 5:52 PM
# Version 1.0
import enum
import json
import os.path

from typing import Dict

from buildfly.utils.compress_utils import uncompress_tar_gz
from buildfly.utils.http_pkg_utils import download_http_pkg
from buildfly.utils.io_utils import write_to_file
from buildfly.utils.log_utils import get_logger
from buildfly.repos.repo_cache_db import repo_cache
from buildfly.utils.github_api_utils import api_client

from collections import namedtuple

from buildfly.utils.system_utils import get_bfly_path

logger = get_logger(__name__)


class BFlyPkg(object):
    def __init__(self, name, group, version):
        self.name = name
        self.group = group
        self.version = version

    def __repr__(self):
        return json.dumps(self.__dict__, indent=2)

    def path(self):
        p = []
        p.append(self.type)
        if self.group:
            p.append(self.group)
        p.append(self.name)
        if self.version:
            p.append("v")
            p.append(self.version)
        return "/".join(p)


class PkgType(enum.Enum):
    GITHUB = 1


class DependencyPlugin(object):
    def tags_info(self, bpkg: BFlyPkg):
        return []

    def list_versions(self, bpkg: BFlyPkg):
        return []

    def download_pkg(self, bpkg: BFlyPkg):
        return


class GithubPlugin(DependencyPlugin):
    def __init__(self, api_client):
        self.api_client = api_client

    def tags_info(self, bpkg: BFlyPkg):
        group = bpkg.group
        if not group:
            group, _ = bpkg.meta['url'].split("/")[-2:]
            bpkg.group = group
        all_tags = self.api_client.list_tags(group, bpkg.name)
        return all_tags[bpkg.version]

    def list_versions(self, bpkg: BFlyPkg):
        ui = bpkg.meta['url'].split("/")
        group, name = ui[-2], ui[-1]
        tags = list(api_client.list_tags(group, name).keys())
        return tags

    def download_pkg(self, bpkg: BFlyPkg):
        logger.info("%s", bpkg)
        tarball_url = bpkg.ver_info['tarball_url']
        commit_sha = bpkg.ver_info['commit']['sha']
        print(tarball_url)

        cache_dir = get_bfly_path(f"cache/{bpkg.path()}", True)
        code_dir = get_bfly_path(f"repo/{bpkg.path()}", True)
        # download
        tmp_pkg_file = f"{cache_dir}/code.tar.gz"
        sha_file = f"{cache_dir}/COMMIT_SHA"
        if os.path.exists(sha_file):
            logger.info(f"{bpkg.path()} tar exist,skip download")
        else:
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

    def parse_pkg(self, name, group, version):
        logger.info("%s %s %s", name, group, version)
        bpkg = BFlyPkg(name, group, version)
        bpkg.meta = repo_cache.get_pkg(name)
        bpkg.type = self.detect_type(bpkg)
        assert bpkg.type in self.plugins, f"plugin {bpkg.type} not support"
        bpkg.ver_info = self.plugins[bpkg.type].tags_info(bpkg)
        return bpkg

    def list_versions(self, bpkg: BFlyPkg):
        return self.plugins[bpkg.type].list_versions(bpkg)

    def download_pkg(self, bpkg: BFlyPkg):
        return self.plugins[bpkg.type].download_pkg(bpkg)

    def build_pkg(self):
        pass

    def install_pkg(self):
        pass


DEP_MGR = DependencyMananger()
DEP_MGR.register("github", GithubPlugin(api_client))

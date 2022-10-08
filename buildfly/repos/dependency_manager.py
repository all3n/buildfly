#!/usr/bin/env python
# -*- coding: utf-8 -*-
# Author wanghuacheng
# Description
# Date 10/8/22 5:52 PM
# Version 1.0
from buildfly.utils.log_utils import get_logger
from buildfly.repos.repo_cache_db import repo_cache
from collections import namedtuple

logger = get_logger(__name__)
BFlyDependency = namedtuple("BFlyDependency", ["name", "group", "version"])


class DependencyMananger(object):
    def __init__(self):
        pass

    def parse_pkg(self, name, group, version):
        logger.info("%s %s %s", name, group, version)
        bdep = BFlyDependency(name, group, version)
        print(bdep)
        rcache = repo_cache.get_pkg(name)

        pass

    def download_pkg(self):
        pass

    def build_pkg(self):
        pass

    def install_pkg(self):
        pass


DEP_MGR = DependencyMananger()

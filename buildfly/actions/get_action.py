#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""

"""
import os

from buildfly.actions.base_action import BaseAction
from buildfly.repos.common import BFlyPkg
from buildfly.utils.log_utils import get_logger
from buildfly.repos.dependency_manager import DEP_MGR

logger = get_logger(__name__)


class GetAction(BaseAction):
    def parse_args(self, parser):
        parser.add_argument('pkg', metavar='pkg', type=str,
                            help="get pkg")

        parser.add_argument('args', metavar='args', type=str, nargs="*",
                            help="cmd args")

    def run(self):
        pkg = self.args.pkg
        args, kwargs = self.split_args()
        self.get_pkg(pkg, kwargs)

    def get_pkg(self, pkg, args=[], kwargs={}):
        group = None
        version = None
        artifact = pkg
        if '@' in pkg:
            artifact, version = pkg.split("@")
        elif 'version' in kwargs:
            version = kwargs["version"]

        if '/' in artifact:
            group, name = artifact.split('/')[-2:]
        else:
            name = artifact
        # name = os.path.basename(group)
        show_versions = False
        if len(args) > 0 and args[0].lower() in ['vers', 'versions', 'ver', 'v']:
            show_versions = True

        bpkg = BFlyPkg(name, group)
        tag = kwargs.get("tag", version)
        if tag:
            bpkg.version = tag
        else:
            commit_sha = kwargs.get("commit_sha", None)
            bpkg.commit_sha = commit_sha

        bpkg = DEP_MGR.parse_pkg(bpkg)
        bpkg.kwargs = kwargs
        if show_versions:
            vers = DEP_MGR.list_versions(bpkg)
            logger.info("%s", vers)
            return

        DEP_MGR.download_pkg(bpkg)
        build_force = kwargs.get("build", "").lower() == "force"
        build_mode = kwargs.get("mode", "Release")
        return DEP_MGR.build_pkg(bpkg, build_force, build_mode)
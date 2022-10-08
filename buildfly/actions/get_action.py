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
from buildfly.config.repo_list import repo_list
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
        if '@' in pkg:
            group,version = pkg.split("@")
        elif 'version' in kwargs:
            group = pkg
            version = kwargs["version"]
        name = os.path.basename(group)

        # names = os.path.basename(pkg).split("@")
        # name = names[0]
        # version = None
        # if len(names) > 1:
        #     version = names[1]
        # elif "version" in kwargs:
        #     version = kwargs["version"]

        # if name in repo_list:
        #     pkg = repo_list.get(name)
            # if version:
            #     pkg = pkg + "@" + version
        # logger.info("%s %s", name, pkg)

        pkg_meta = DEP_MGR.parse_pkg(name, group, version)
        # app_dep = BuildDependency(name, pkg)
        # get_dep(app_dep)

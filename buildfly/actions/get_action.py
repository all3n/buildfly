#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""

"""
from buildfly.actions.base_action import BaseAction
from buildfly.utils.dep_utils import *
from buildfly.utils.yaml_conf_utils import BuildDependency
from buildfly.config.repo_list import repo_list
import os
import logging

logger = logging.getLogger(__name__)


class GetAction(BaseAction):
    def parse_args(self, parser):
        parser.add_argument('pkg', metavar='pkg', type=str,
                            help="get pkg")

    def run(self):
        pkg = self.args.pkg
        names = os.path.basename(pkg).split("@")
        name = names[0]
        version = None
        if len(names) > 1:
            version = names[1]
        if name in repo_list:
            pkg = repo_list.get(name)
            if version:
                pkg = pkg + "@" + version
        logger.info("%s %s", name, pkg)
        app_dep = BuildDependency(name, pkg)
        get_dep(app_dep)

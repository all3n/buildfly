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
from buildfly.config.global_config import G_CONFIG
from buildfly.utils.system_utils import get_bfly_path
import os

import logging

logger = logging.getLogger(__name__)


class RepoAction(BaseAction):

    def parse_args(self, parser):
        parser.add_argument('cmd', metavar='cmd', type=str, nargs="?",
                            help="repo cmd")

    def run(self):
        cmd = self.args.cmd

        repos = G_CONFIG.get_value("repos")
        for repo in repos:
            rtype = type(repo)
            if rtype == str and repo.startswith("/"):
                print(f"local repo {repo}")
            else:
                name = repo["name"]
                git = repo["git"]
                local_repo_path = get_bfly_path(os.path.join("repos", name))
                if not os.path.exists(local_repo_path):
                    os.makedirs(local_repo_path)
                os.system(f"git clone {git} {local_repo_path}")




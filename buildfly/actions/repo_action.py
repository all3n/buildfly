#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2022 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.


from buildfly.actions.base_action import BaseAction, SubCmdAction
from buildfly.config.global_config import G_CONFIG
from buildfly.utils.log_utils import get_logger
from buildfly.utils.system_utils import get_bfly_path
import os

logger = get_logger(__name__)
from buildfly.repos.repo_cache_db import repo_cache


def repo_filter(f):
    return f not in ["README.md", ".gitignore", ".git"]


class RepoAction(SubCmdAction):
    def cmd_cache(self, name="", repo_path=None):
        logger.info(f"gen repo cache {name}")

        def get_lib_dirs(d, base):
            ret = []
            if os.path.isdir(os.path.join(base, d)):
                if os.path.exists(os.path.join(base, d, 'manifest.json')):
                    ret.append(d)
                else:
                    for sd in os.listdir(os.path.join(base, d)):
                        ret.extend(get_lib_dirs(os.path.join(d, sd), base))
            return ret

        if name:
            if repo_path is None:
                local_repo_path = get_bfly_path(os.path.join("repos", name))
                if not os.path.exists(local_repo_path):
                    logger.warn(f"{local_repo_path} not exists")
                    return
                repo_dirs = get_lib_dirs("", local_repo_path)
            else:
                local_repo_path = repo_path
                repo_dirs = get_lib_dirs("", local_repo_path)
            repo_cache.update_repo(name, local_repo_path, repo_dirs)

    def get_repos(self):
        return {r["name"]: r for r in G_CONFIG.get_value("repos")}

    def cmd_sync(self):
        repos = self.get_repos()
        for name, repo in repos.items():
            path = repo["path"]
            name = repo["name"]
            if path.startswith("/"):
                logger.info(f"{path} local repo")
                local_repo_path = path
            else:
                local_repo_path = get_bfly_path(os.path.join("repos", name))
                if not os.path.exists(local_repo_path):
                    gcmd = f"git clone {path} {local_repo_path}"
                    logger.info(gcmd)
                    os.system(gcmd)
                else:
                    gcmd = f"cd {local_repo_path};git pull"
                    logger.info(gcmd)
                    os.system(gcmd)
            self.cmd_cache(name, local_repo_path)

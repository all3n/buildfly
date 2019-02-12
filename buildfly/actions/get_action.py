#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""

"""
from buildfly.actions.basic_action import basic_action
from buildfly.utils.compress_utils import *
from buildfly.utils.http_pkg_utils import download_http_pkg
import os
import sys
from collections import namedtuple
from buildfly.build.build_manager import build_manager
class get_action(basic_action):
    def parse_args(self, parser):
        parser.add_argument('libdesc', metavar='libdesc', type=str,
                    help="lib description")
    def run(self):
        lib_info = self.detact_lib_type(self.args.libdesc)
        if lib_info:
            lib_type = lib_info["type"]
            if lib_type == "github":
                repo_desc = ("github", lib_info["owner"], lib_info["repo"],lib_info["repo_info"][0])
                repo_dir_path = os.path.expanduser("~/.buildfly/repo/%s/%s/%s/%s" % repo_desc)
                codepath = self.process_github_lib(lib_info, repo_dir_path)
            elif lib_type == "git":
                repo_dir_path = os.path.expanduser("~/.buildfly/repo/%s" % lib_info["url"])
                codepath = self.git_clone_lib_src(lib_info, repo_dir_path)
            self.build_code_and_install(codepath, lib_info)


    def detact_lib_type(self, libdesc):
        # github.com
        lib_type_info = {}
        if ":" not in libdesc:
            libinfo = libdesc.split("/")
            if len(libinfo) == 2:
                lib_type_info["type"] = "github"
                lib_type_info["owner"] = libinfo[0]
                lib_type_info["repo"] = libinfo[1]
                if "@" in libinfo[1]:
                    # owner/repo@tags
                    lib_repo_info = libinfo[1].split("@")
                    lib_type_info["repo"] = lib_repo_info[0]
                    lib_type_info["repo_info"] = ("tag", lib_repo_info[1])
                else:
                    # owner/repo
                    lib_type_info["repo_info"] = ("branch", "master")
            else:
                raise Exception("github lib must have owner/repo")
        elif libdesc.endswith(".git"):
            lib_type_info["type"] = "git"
            lib_type_info["url"] = libdesc
        else:
            raise Exception("desc type not support")
        return lib_type_info

    def process_github_lib(self, lib_info):
        from buildfly.utils.github_api_utils import api_client
        repo_type,repo_val = lib_info["repo_info"]
        if repo_type == "tag":
            tags_info = api_client.list_tags(lib_info["owner"],lib_info["repo"])
            if repo_val in tags_info:
                tarball_url = tags_info[repo_val]["tarball_url"]
                repo_desc = ("github", lib_info["owner"], lib_info["repo"],lib_info["repo_info"][0])
                tmp_file_path = os.path.expanduser("~/.buildfly/cache/%s/%s/%s/%s.tar.gz" % repo_desc)
                download_http_pkg(tarball_url,tmp_file_path)
                uncompress_tar_gz(repo_dir_path, tmp_file_path)
                subdirs = os.listdir(repo_dir_path)
                dircnt = len(subdirs)
                if dircnt == 1:
                    codepath = os.path.join(repo_dir_path, subdirs[0])
                    if not os.path.isdir(codepath):
                        raise Exception("if code directory has one file,this file must be directory")
                elif dircnt == 0:
                    raise Exception("code dir must not empty")
                else:
                    codepath = repo_dir_path

                print("get repo: %s" % (repo_dir_path))
            else:
                print("%s tag is not valid" % repo_val)
                print("tag list: %s" % (list(tags_info.keys())))
                sys.exit(-1)

    def build_code_and_install(self, code_dir, repo_desc):
        bm = build_manager()
        print("start build %s" % code_dir)
        bm.build(code_dir, repo_desc)

    def git_clone_lib_src(self, lib_info, repo_dir_path):
        CMD = "git clone --recursive %s %s" % (lib_info["url"], repo_dir_path)
        os.system(CMD)






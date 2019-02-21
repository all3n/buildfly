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
import sys
from collections import namedtuple
from buildfly.build.build_manager import build_manager
from buildfly.utils.compress_utils import *
from buildfly.utils.http_pkg_utils import download_http_pkg
from buildfly.utils.github_api_utils import api_client

def get_bfly_dir(d):
    return os.path.expanduser("~/.buildfly/%s" % d)

def get_repo_level_dir(lib_info, category = "repo"):
    return get_bfly_dir("{category}/{lib_type}/{owner}/{repo}/{repo_info}".format(
            category = category,
            lib_type = lib_info["type"],
            owner = lib_info['owner'],
            repo = lib_info['repo'],
            repo_info = "/".join(lib_info['repo_info'])
        ))

def get_bfly_cache_dir(lib_info):
    return get_repo_level_dir(lib_info, "cache")

def get_bfly_repo_dir(lib_info):
    return get_repo_level_dir(lib_info, "code")

def get_blfy_install_dir(lib_info):
    return get_repo_level_dir(lib_info, "install")


def get_dep(libdesc):
    lib_info = detact_lib_type(libdesc)
    if lib_info:
        install_dir = get_blfy_install_dir(lib_info)
        if os.path.exists(install_dir):
            print("%s exists in %s" % (libdesc, install_dir))
            return
        lib_type = lib_info["type"]
        if lib_type == "github":
            repo_dir_path = get_bfly_repo_dir(lib_info)
            process_github_lib(lib_info, repo_dir_path)
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
        elif lib_type == "git":
            codepath = repo_dir_path
            repo_dir_path = get_bfly_repo_dir("%s" % lib_info["url"])
            git_clone_lib_src(lib_info, codepath)
        build_code_and_install(codepath, install_dir)


def detact_lib_type(libdesc):
    # github.com
    lib_info = {}
    if ":" not in libdesc and "/" in libdesc:
        libdesc_info = libdesc.split("/")
        if len(libdesc_info) == 2:
            owner, repo = libdesc_info
            lib_info["type"] = "github"
            lib_info["owner"] = owner
            repo_info = repo.split("@")
            repo_name = repo_info[0]
            lib_info["repo"] = repo_name

            if not api_client.is_repo_exists(owner, repo_name):
                print("http://github.com/%s/%s repo not exists" % (owner, repo_name))
                sys.exit(-1)

            # has tags
            if len(repo_info) > 1:
                # owner/repo@tags
                repo_tag = repo_info[1]
                lib_info["repo"] = repo_name
                lib_info["repo_info"] = ["tag", repo_tag]
                tags_info = api_client.list_tags(owner, repo_name)
                if repo_tag in tags_info:
                    tarball_url = tags_info[repo_tag]["tarball_url"]
                    lib_info['tarball_url'] = tarball_url
                else:
                    print("%s tag is not valid" % repo_tag)
                    print("tag list: %s" % (list(tags_info.keys())))
                    sys.exit(-1)
            else:
                # branch
                default_branch = "master"
                branch_info = api_client.get_branch_info(owner, repo, default_branch)
                commit_sha = branch_info["commit"]['sha']
                lib_info["repo_info"] = ["branch", default_branch, commit_sha]
                tarball_url = "https://github.com/%s/%s/archive/%s.tar.gz" % (owner, repo, default_branch)
                lib_info['tarball_url'] = tarball_url
        else:
            print("github lib must match pattern owner/repo")
            sys.exit(-1)

    elif libdesc.endswith(".git"):
        lib_info["type"] = "git"
        lib_info["url"] = libdesc
    else:
        DESC_TYPE_HELP="""
        only support
            onwer/repo          github repo path
            xxxxx.git           git url,will use local git clone code from this url
        """
        print("%s is not valid repo pattern,not support:%s" % (libdesc, DESC_TYPE_HELP))
        sys.exit(-1)
    return lib_info

def process_github_lib(lib_info, repo_dir_path):
    print(lib_info)
    cache_dir = get_bfly_cache_dir(lib_info)
    tmp_file_path = os.path.join(cache_dir, "code.tar.gz")
    download_http_pkg(lib_info['tarball_url'],tmp_file_path)
    uncompress_tar_gz(repo_dir_path, tmp_file_path)


def build_code_and_install(code_dir, lib_info):
    bm = build_manager()
    print("start build %s" % code_dir)
    bm.build(code_dir, lib_info)

def git_clone_lib_src(lib_info, repo_dir_path):
    CMD = "git clone --recursive %s %s" % (lib_info["url"], repo_dir_path)
    os.system(CMD)




def get_dep_compile_options(libdesc):
    lib_info = detact_lib_type(libdesc)
    install_lib_dir = get_blfy_install_dir(lib_info)
    dirs = os.listdir(install_lib_dir)
    libname = ""
    compile_options = []
    for d in dirs:
        abs_dir = os.path.join(install_lib_dir, d)
        if d == "lib" or d == "lib64":
            pkgconfig_dir = os.path.join(install_lib_dir, "lib", "pkgconfig")
            if False and os.path.exists(pkgconfig_dir):
                # if pkgconfig exists,use pkgconfig cflags
                cmd = "PKG_CONFIG_PATH=%s:$PKG_CONFIG_PATH pkg-config --libs --cflags %s" % (pkgconfig_dir,libname)
                pkgconfig_cflags = os.popen(cmd).read().rstrip()
                return pkgconfig_cflags
            else:
                compile_options.append("-L%s" % abs_dir)
                # compile_options.append("-l%s" % libname)
        elif d == "include":
            compile_options.append("-I%s" % abs_dir)
    return " ".join(compile_options)







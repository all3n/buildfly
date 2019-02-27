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
from buildfly.utils.io_utils import *
from buildfly.utils.http_pkg_utils import download_http_pkg
from buildfly.utils.github_api_utils import api_client
from buildfly.utils.system_utils import exec_cmd
import logging
from collections import namedtuple
DevOptions = namedtuple("DepOptions", ["cflags", "libs", "libs_path", "libs_path_option", "libs_other"])


COMMIT_SHA=".COMMIT_SHA"

def get_dep(app_dep):
    lib_info = app_dep.lib_info
    install_dir = app_dep.get_install_dir()
    cache_dir = app_dep.get_cache_dir()
    repo_dir = app_dep.get_repo_dir()
    if lib_info:
        if os.path.exists(install_dir):
            print("%s exists in %s" % (app_dep.name, install_dir))
            return
        lib_type = lib_info["type"]
        if lib_type == "github":
            repo_dir_path = app_dep.get_repo_dir()
            process_github_lib(app_dep, repo_dir_path)
            code_dir = app_dep.get_code_dir()
        elif lib_type == "git":
            code_dir = repo_dir_path
            repo_dir_path = app_dep.get_repo_dir()
            git_clone_lib_src(lib_info, code_dir)
    elif app_dep.dep_type == 'url':
        if os.path.exists(install_dir):
            print("%s exist" % app_dep.name)
            return
        tmp_file_path = os.path.join(cache_dir, "code.tar.gz")
        if not os.path.exists(tmp_file_path):
            download_http_pkg(app_dep.url,tmp_file_path)
        uncompress_tar_gz(repo_dir, tmp_file_path)
        code_dir = app_dep.get_code_dir()


    bm = build_manager()
    bm.build(app_dep)


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
                lib_info['url'] = "https://github.com/%s/%s/tree/%s" % (owner, repo_name, repo_tag)
                tags_info = api_client.list_tags(owner, repo_name)
                if repo_tag in tags_info:
                    tag_info = tags_info[repo_tag]
                    #print(tag_info)
                    tarball_url = tag_info["tarball_url"]
                    lib_info['tarball_url'] = tarball_url
                    lib_info['commit'] = tag_info["commit"]
                else:
                    print("%s tag is not valid" % repo_tag)
                    print("tag list: %s" % (list(tags_info.keys())))
                    sys.exit(-1)
            else:
                # branch
                default_branch = "master"
                lib_info['url'] = "https://github.com/%s/%s/tree/%s" % (owner, repo_name, default_branch)
                branch_info = api_client.get_branch_info(owner, repo, default_branch)
                commit_info = branch_info["commit"]
                lib_info['commit'] = commit_info
                commit_sha = commit_info['sha']
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

def process_github_lib(app_dep, repo_dir_path):
    lib_info = app_dep.lib_info
    commit_sha = lib_info['commit']['sha']
    cache_dir = app_dep.get_cache_dir()
    tmp_file_path = os.path.join(cache_dir, "code.tar.gz")
    commit_sha_file = os.path.join(cache_dir, COMMIT_SHA)

    if read_file_line(commit_sha_file) == commit_sha:
        logging.info("%s code exist" % commit_sha)
    else:
        download_http_pkg(lib_info['tarball_url'],tmp_file_path)
        write_to_file(commit_sha_file, commit_sha)

    uncompress_tar_gz(repo_dir_path, tmp_file_path)


def git_clone_lib_src(lib_info, repo_dir_path):
    CMD = "git clone --recursive %s %s" % (lib_info["url"], repo_dir_path)
    os.system(CMD)


def get_dep_compile_options(app_dep, dep_libs):
    lib_info = app_dep.lib_info
    install_lib_dir = app_dep.get_install_dir()
    dirs = os.listdir(install_lib_dir)
    cflags = []
    libs = []
    lib_dirs = []
    for d in dirs:
        abs_dir = os.path.join(install_lib_dir, d)
        if d == "lib" or d == "lib64":
            pkgconfig_dir = os.path.join(install_lib_dir, "lib", "pkgconfig")
            lib_names = [l.lib_name for l in dep_libs]
            if os.path.exists(pkgconfig_dir):
                # if pkgconfig exists,use pkgconfig cflags
                pkg_prefix="PKG_CONFIG_PATH=%s:$PKG_CONFIG_PATH pkg-config"
                cmd = "%s --cflags %s" % (pkg_prefix, pkgconfig_dir," ".join(lib_names))
                pkgconfig_cflags = exec_cmd(cmd)

                cmd = "%s --libs-only-L %s" % (pkg_prefix, pkgconfig_dir," ".join(lib_names))
                pkgconfig_libs_L_path_option = exec_cmd(cmd)
                pkgconfig_libs_L_path = pkgconfig_libs_L_path_option.replace("-L","").split(" ")

                cmd = "%s --libs-only-l %s" % (pkg_prefix, pkgconfig_dir," ".join(lib_names))
                pkgconfig_libs_l = exec_cmd(cmd)

                cmd = "%s --libs-only-other %s" % (pkg_prefix, pkgconfig_dir," ".join(lib_names))
                pkgconfig_libs_other = exec_cmd(cmd)
                dev_options = DevOptions(cflags=pkgconfig_cflags, libs = pkgconfig_libs_l,
                        libs_path = pkgconfig_libs_L_path,
                        libs_path_option = pkgconfig_libs_L_path_option,
                        libs_other = pkgconfig_libs_other
                        )

                return dev_options
            else:
                lib_dirs.append(abs_dir)
                libs += lib_names
        elif d == "include":
            cflags.append("-I%s" % abs_dir)
        else:
            pass
    dev_options = DevOptions(cflags = " ".join(cflags),
                libs = " ".join(["-l%s" % ln for ln in libs]),
                libs_path = lib_dirs,
                libs_path_option = " ".join(["-L%s" % ln for ln in lib_dirs]),
                libs_other = None
                )

    return dev_options







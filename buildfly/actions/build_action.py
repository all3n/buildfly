#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""

"""
import sys
import os
import glob
import json
import re
from buildfly.actions.basic_action import basic_action
from buildfly.utils.yaml_conf_utils import yaml_conf_loader
from buildfly.utils.dep_utils import *
from buildfly.utils.system_utils import *

CONF_NAME="buildfly.yaml"
COMPILER_PATTERN = re.compile("(\w+)([<=>]{1,2})?([\d\.\w]+)?")


class build_action(basic_action):
    def parse_args(self, parser):
        parser.add_argument('target', metavar='target', type=str, nargs = "?",
                    help="build target")

    def run(self):
        cur_dir = os.path.abspath(sys.path[0])
        print("run build action in %s" % (cur_dir))
        self.parse_build_conf(os.path.join(cur_dir, CONF_NAME))
        self.check_compiler(self.compiler_info)
        self.start_build()

    def parse_build_conf(self, conf_file):
        if not os.path.exists(conf_file):
            print("%s not exist!" % conf_file)
            sys.exit(-1)

        app_conf = yaml_conf_loader(conf_file)
        print(app_conf)
        self.app_conf = app_conf

        self.build_dir = app_conf.args.get('build_dir', 'build')
        if not os.path.exists(self.build_dir):
            os.makedirs(self.build_dir)

        self.bins = app_conf.bins
        self.libs = app_conf.libs


        compiler = app_conf.args.get('compiler', 'cmake').replace(" ","")
        self.compiler_info = COMPILER_PATTERN.match(compiler).groups()

    def check_compiler(self, compiler_info):
        print(compiler_info)
        # check local
        compiler_name,version_op, version = compiler_info
        version_match = False
        if compiler_name == "cmake":
            if check_command_exists(compiler_name):
                verson_str = exec_cmd("cmake --version")
                version_res = re.search("version ([\d\.]+)", verson_str)
                if version_res:
                    compiler_version = version_res.group(1)
                    print("compiler:system:%s:%s" % (compiler_name, compiler_version))
                    if version_op and version:
                        if (version_op == "=" or version_op == "==") and version == compiler_version:
                            version_match = True
                        elif version_op == ">":
                            if compiler_version > version:
                                version_match = True
                        elif version_op == ">=":
                            if compiler_version >= version:
                                version_match = True
                        elif version_op == "<":
                            if compiler_version < version:
                                version_match = True
                        elif version_op == "<=":
                            if compiler_version <= version:
                                version_match = True
        if version_match:
            print("compiler: version ok")
        else:
            print("compiler: version not match %s%s%s" % compiler_info)










    def start_build(self):
        self.build_flag = {}
        target = self.args.target

        if target:
            if target in self.bins:
                self.build_bin(target, self.bins[target])
            if target in self.libs:
                self.build_library(target, self.libs[target])

        else:
            for name, build_info in self.bins.items():
                self.build_bin(name, build_info)

            for name, build_info in self.libs.items():
                if name not in self.build_flag:
                    self.build_library(name, build_info)

    def get_dep_raw(dep_key):
        dep_name = dep.split(":")[0]

        if dep_name in self.app_conf.dependency:
            return self.app_conf.dependency[dep_name]
        else:
            print("%s not defined" % dep_name)
            sys.exit(-1)

    def build_dep(self, name, build_info):
        if 'deps' not in build_info:
            return

        print(name, json.dumps(build_info, indent=2))
        deps = build_info['deps']
        for dep in deps:
            dep_name = self.get_dep_raw(dep)
            if dep_name not in self.build_flag:
                if dep.startswith("//"):
                    dep = dep[2:]
                    if self.build_library(dep, self.libs[dep]):
                        self.build_flag[dep] = True
                    else:
                        raise Exception("build dep library:%s fail" % (dep))
                else:
                    print("dep:%s" % dep)
                    get_dep(dep)

    def expand_pattern(self, pattern):
        return glob.glob(pattern, recursive = True)


    def build_bin(self, name, build_info):
        bin_dir = os.path.join(self.build_dir, "build-bin-%s" % name)
        if not os.path.exists(bin_dir):
            os.makedirs(bin_dir)
        self.build_dep(name, build_info)
        cmds = []
        srcs = build_info['srcs']
        libs = build_info['libs']
        includes = build_info['includes'] if 'includes' in build_info else []
        target = name
        library_path = []
        link_library = []
        static_libs = []
        dep_options = []
        if "deps" in build_info:
            deps = build_info['deps']
            for dep in deps:
                if dep.startswith("//"):
                    dep = dep[2:]
                    lib_build_dir = os.path.join(self.build_dir, "build-lib-%s" % dep)
                    dep_lib_info = self.libs[dep]
                    lib_include_dir = dep_lib_info['includes']
                    lib_type = dep_lib_info.get('lib_type', 'shared')
                    includes += lib_include_dir
                    if lib_type == 'shared':
                        library_path.append(lib_build_dir)
                        link_library.append(dep)
                    else:
                        static_libs.append(os.path.join(lib_build_dir, "lib%s.a" % (dep)))
                else:
                    dep_options.append(get_dep_compile_options(dep))

        link_library += libs
        include_options = " ".join(["-I%s" % i for i in includes])
        library_path_options = " ".join(["-L%s" % i for i in library_path])
        link_lib_options = " ".join(["-l%s" % i for i in link_library])
        static_lib_option=" ".join(static_libs)
        dep_extra_lib_option=" ".join(dep_options)

        srcs_files = []
        for src in srcs:
            srcs_files += self.expand_pattern(src)
        srcs_options = " ".join(srcs_files)
        cflags_options = build_info.get('cflags', '')

        cmds.append("g++ {cflags} {dep_extra_lib_option} {library_path_options} {link_lib_options} {include_options} -o {build_dir}/{target} {srcs} {static_lib_option}".format(
            target = target,
            srcs = srcs_options,
            library_path_options = library_path_options,
            include_options = include_options,
            build_dir = bin_dir,
            cflags = cflags_options,
            link_lib_options=link_lib_options,
            static_lib_option=static_lib_option,
            dep_extra_lib_option = dep_extra_lib_option
        ))

        for cmd in cmds:
            print(cmd)
            os.system(cmd)
        return True

    def build_library(self, name, build_info):
        lib_dir = os.path.join(self.build_dir, "build-lib-%s" % name)
        if not os.path.exists(lib_dir):
            os.makedirs(lib_dir)
        self.build_dep(name, build_info)

        cmds = []
        srcs = build_info['srcs']
        includes = build_info['includes'] if 'includes' in build_info else []
        target = name
        include_options = " ".join(["-I%s" % i for i in includes])

        srcs_files = []
        for src in srcs:
            srcs_files += self.expand_pattern(src)
        srcs_options = " ".join(srcs_files)
        cflags_options = build_info.get('cflags', '')
        src_names = [i.rsplit(".",1)[0] for i in srcs_files]

        for sn in src_names:
            object_file = "%s/%s.o" % (lib_dir, sn)
            object_file_dir = os.path.dirname(object_file)
            if not os.path.exists(object_file_dir):
                os.makedirs(object_file_dir)
            cmds.append("g++ -c {cflags} {include_options} -o {object_file} {srcs}".format(
                object_file=object_file,
                srcs = srcs_options,
                include_options = include_options,
                cflags = cflags_options
            ))
        lib_type = build_info.get('lib_type', 'shared')
        lib_objects_option = "".join(["%s/%s.o" % (lib_dir,i) for i in src_names])
        if lib_type == 'shared':
            target_ext='so'
            cmds.append("g++ -shared -fPIC -o {build_dir}/lib{target}.{target_ext} {lib_objects_option}".format(
                    target=target,
                    lib_objects_option = lib_objects_option,
                    target_ext=target_ext,
                    build_dir=lib_dir
                ))
        else:
            target_ext='a'
            cmds.append("ar crv {build_dir}/lib{target}.{target_ext} {lib_objects_option}".format(
                    target=target,
                    lib_objects_option = lib_objects_option,
                    target_ext=target_ext,
                    build_dir=lib_dir
                ))

        for cmd in cmds:
            print(cmd)
            os.system(cmd)

        return True


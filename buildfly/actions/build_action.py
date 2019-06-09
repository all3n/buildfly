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
import logging
import stat
import re

from buildfly.actions.basic_action import basic_action
#from buildfly.utils.yaml_conf_utils import yaml_conf_loader
from buildfly.utils.dep_utils import *
from buildfly.utils.color_utils import *
from buildfly.utils.system_utils import *
from buildfly.targets import *

CONF_NAME="buildfly.yaml"
BFLY_CONF = "BUILDFLY"
COMPILER_PATTERN = re.compile("(\w+)([<=>]{1,2})?([\d\.\w]+)?")
LIBC_VERSION_PATTERN = re.compile("libc-([\d\.]+)\.so")


class build_action(basic_action):
    def parse_args(self, parser):
        parser.add_argument('target', metavar='target', type=str, nargs = "?",
                    help="build target")
        parser.add_argument("-b",'--build_dir',  type=str, default="build",
                            help="build target")


    def run(self):
        cur_dir = os.path.abspath(sys.path[0])

        self.build_dir = self.args.build_dir
        if not os.path.exists(self.build_dir):
            os.makedirs(self.build_dir)
        self.build_dir = os.path.abspath(self.build_dir)
        BUILD_MANAGER.src_dir = cur_dir
        BUILD_MANAGER.build_dir = self.build_dir


        # print("run build action in %s" % (cur_dir))
        bfly_conf = os.path.join(cur_dir, BFLY_CONF)
        export_vars = BUILD_MANAGER.get_exports()
        try:
            if os.path.exists(bfly_conf):
                self.load_bfly_conf(cur_dir, "", export_vars)
            else:
                raise Exception("%s not exists" % BFLY_CONF)
        except NameError as e:
            support_keys = set(export_vars.keys())
            support_keys.remove("__builtins__")
            print("%s \nonly support: %s " % (e, support_keys))
            sys.exit(-1)

        BUILD_MANAGER.check_deps_valid()
        print("build rules load finished!")

        BUILD_MANAGER.run(self.args.target)

        #self.parse_build_conf(os.path.join(cur_dir, CONF_NAME))
        #self.check_compiler(self.compiler_info)
        # self.check_glibc()
        #self.start_build()

    def load_bfly_conf(self, dir, package, export_vars):
        res = {"package": package}
        # TODO
        BUILD_MANAGER.package = "//" + package
        bconf = os.path.join(dir, BFLY_CONF)
        bfly_exec(bconf, export_vars, res)
        for subp in os.listdir(dir):
            subp_file = os.path.join(dir, subp)
            subp_conf = os.path.join(subp_file, BFLY_CONF)
            if os.path.isdir(subp) and os.path.isfile(subp_conf):
                sub_package = package + "/" + subp if package else subp
                self.load_bfly_conf(subp_file, sub_package, export_vars)
        BUILD_MANAGER.package = None





    def parse_build_conf(self, conf_file):
        if not os.path.exists(conf_file):
            print("%s not exist!" % conf_file)
            sys.exit(-1)

        #app_conf = yaml_conf_loader(conf_file)
        app_conf = None
        # print(app_conf)
        self.app_conf = app_conf

        self.build_dir = app_conf.args.get('build_dir', 'build')
        if not os.path.exists(self.build_dir):
            os.makedirs(self.build_dir)

        self.bins = app_conf.bins
        self.libs = app_conf.libs


        compiler = app_conf.args.get('compiler', 'cmake').replace(" ","")
        self.compiler_info = COMPILER_PATTERN.match(compiler).groups()

    def check_glibc(self):
        libc_so_file = exec_cmd("readlink -f `ldconfig -p|grep libc.so.6|head -n 1|awk -F\"=> \" '{print $2}'`")
        version_match = LIBC_VERSION_PATTERN.search(libc_so_file)
        if version_match:
            libc_version = version_match.group(1)
            logging.info("libc version:%s" % libc_version)
        required_libc = str(self.app_conf.args.get("glibc", ""))
        logging.info("requried glibc:%s" % required_libc)
        if libc_version and required_libc:
            if required_libc == libc_version:
                logging.info("libc version must==%s,system is %s" % (required_libc, libc_version))
                sys.exit(-2)
            else:
                get_glibc(required_libc)



    def check_compiler(self, compiler_info):
        logging.info(green("check compiler info ") + str(compiler_info))
        # check local
        compiler_name,version_op, version = compiler_info
        version_match = False
        if compiler_name == "cmake":
            if check_command_exists(compiler_name):
                verson_str = exec_cmd("cmake --version")
                version_res = re.search("version ([\d\.]+)", verson_str)
                if version_res:
                    compiler_version = version_res.group(1)
                    # print("compiler:system:%s:%s" % (compiler_name, compiler_version))
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
            logging.info(green("compiler: version ok"))
        else:
            logging.error(red("compiler: version not match %s%s%s" % compiler_info))










    def start_build(self):
        self.build_flag = {}
        target = self.args.target

        if target:
            if target in self.bins:
                self.build_bin(target, self.bins[target])
            elif target in self.libs:
                self.build_library(target, self.libs[target])

        else:
            for name, build_info in self.bins.items():
                self.build_bin(name, build_info)

            for name, build_info in self.libs.items():
                if name not in self.build_flag:
                    self.build_library(name, build_info)


    def build_dep(self, name, build_info):
        if 'deps' not in build_info:
            return
        deps = build_info['deps']
        # print(deps)
        for dep_name, dep_libs in deps.items():
            for dep_lib in dep_libs:
                # print("%s:%s" %  (dep_name, dep_lib))
                if dep_lib.name not in self.build_flag:
                    # print(dep_lib)
                    if dep_lib.libdesc.startswith("//"):
                        if self.build_library(dep_lib.name, self.libs[dep_lib.name]):
                            self.build_flag[dep_lib.name] = True
                        else:
                            raise Exception("build dep library:%s fail" % (dep_lib))
                    else:
                        app_dep = self.app_conf.dependency[dep_lib.name]
                        get_dep(app_dep)

    def expand_pattern(self, pattern):
        return glob.glob(pattern, recursive = True)


    def process_dep_options(self, coption, dep_libs):
        dep_options = []
        dep_options.append(coption.cflags)
        dep_options.append(coption.libs_path_option)
        dep_options.append(coption.libs_option)
        if coption.libs_other:
            dep_options.append(coption.libs_other)
        return dep_options

    def build_bin(self, name, build_info):
        bin_dir = os.path.join(self.build_dir, "build-bin-%s" % name)
        if not os.path.exists(bin_dir):
            os.makedirs(bin_dir)
        self.build_dep(name, build_info)
        cmds = []
        srcs = build_info['srcs']
        libs = build_info.get('libs', [])
        includes = build_info['includes'] if 'includes' in build_info else []
        target = name
        library_path = []
        link_library = []
        static_libs = []
        dep_options = []

        dep_all_lib_dirs = []
        if "deps" in build_info:
            deps = build_info['deps']
            print("deps:%s" % deps)
            for dep_name, dep_libs in deps.items():
                for dep_lib in dep_libs:
                    if dep_lib.libdesc.startswith("//"):
                        print(dep_lib)
                        dep = dep_lib.name
                        lib_build_dir = os.path.join(self.build_dir, "build-lib-%s" % dep)
                        dep_lib_info = self.libs[dep]
                        lib_include_dir = dep_lib_info['includes']
                        lib_type = dep_lib.link_type
                        includes += lib_include_dir
                        if lib_type == 'shared':
                            library_path.append(lib_build_dir)
                            link_library.append(dep)
                        else:
                            static_libs.append(os.path.join(lib_build_dir, "lib%s.a" % (dep)))
                    else:
                        app_dep = self.app_conf.dependency[dep_name]
                        coption = get_dep_compile_options(app_dep , dep_libs)
                        dep_options += self.process_dep_options(coption, dep_libs)
                        dep_all_lib_dirs += coption.share_libs_path

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

        cmds.append("g++ {cflags} {library_path_options} {link_lib_options} \
{include_options} -o {build_dir}/{target} {srcs} \
{dep_extra_lib_option} {static_lib_option}".format(
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
            logging.info(cyan(cmd))
            status_code = os.system(cmd)
            if status_code != 0:
                logging.info("%s exec fail:%d" % (red(cmd), status_code))
                sys.exit(status_code)

        # create run.sh for test bin
        run_bin_script = os.path.join(bin_dir, "run.sh")
        with open(run_bin_script, "w") as rf:
            rf.write("#!/usr/bin/env bash\n")
            rf.write("bin=`dirname \"$0\"`\n")
            rf.write("export APP_DIR=`cd \"$bin/\"; pwd`\n")

            if dep_all_lib_dirs:
                rf.write("# set library path\n")
                rf.write("export LD_LIBRARY_PATH=%s:$LD_LIBRARY_PATH\n" % (":".join(dep_all_lib_dirs)))
            rf.write("# run %s\n" % target)
            rf.write("$APP_DIR/%s $@\n" % (target))
        os.chmod(run_bin_script, stat.S_IRWXU | stat.S_IRWXG | stat.S_IROTH | stat.S_IWOTH)
        logging.info("you can run\n\t%s\nfor test" % red(run_bin_script))



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
            logging.info(cyan(cmd))
            os.system(cmd)

        return True


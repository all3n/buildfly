from buildfly.generator.base_generator import base_generator

from buildfly.utils.system_utils import target_ext
import os


class makefile_generator(base_generator):

    def write_common(self):
        COMMON = """\
CC=gcc
CXX=g++
CCFLAGS=-fPIC
CFLAGS=-fPIC
SRC_DIR={SRC_DIR}
BUILD_DIR={BUILD_DIR}
BIN_DIR=$(BUILD_DIR)/bin
OBJ_DIR=$(BUILD_DIR)/obj
LIB_DIR=$(BUILD_DIR)/lib

$(OBJ_DIR)/%.o : $(SRC_DIR)/%.c
\tmkdir -p $(shell dirname $@)
\t$(CC) -c $< $(CCFLAGS) -o  $@ 

$(OBJ_DIR)/%.o : $(SRC_DIR)/%.cc
\tmkdir -p $(shell dirname $@)
\t$(CXX) -c $< $(CCFLAGS) -o  $@ 

$(OBJ_DIR)/%.o : $(SRC_DIR)/%.cpp
\tmkdir -p $(shell dirname $@)
\t$(CXX) -c $< $(CCFLAGS) -o  $@ 


$(OBJ_DIR):
\tmkdir -p $@ 

""".format(SRC_DIR=self.src_dir, BUILD_DIR=self.build_dir)
        self.write_line(COMMON)

    def convert_src_obj(self, srcs, package):
        SRC_BASE = package.rsplit(":", 1)[0].replace("//", "$(OBJ_DIR)/")
        return ",".join([SRC_BASE + "/" + src.rsplit(".", 1)[0] + ".o" for src in srcs])

    def convert_opts(self, opts, prefix=""):
        if not opts:
            return ""

        return " ".join([prefix + opt for opt in opts])

    def convert_label(self, label):
        if not ":" in label:
            return label
        package, name = label.split(":")
        return self.convert_package_name(package, name)

    def convert_package_name(self, package, name):
        package = package.replace("//", "")
        if package:
            return package.replace("/", "_") + "_" + name
        else:
            return name

    def convert_deps(self, deps):
        if not deps:
            return ""
        else:
            return " ".join([self.convert_label(dep) for dep in deps])

    def convert_package_build_path(self, package, base_path="$(BUILD_DIR)/"):
        package_path = package.replace("//", base_path)
        return package_path

    def gen(self, targets):
        target_list = []
        package_bin_set = set()

        share_ext, static_ext, exec_ext = target_ext()

        for label, target in targets.items():
            is_share = target.is_shared()
            is_lib = target.target_type == "lib"
            if is_lib:
                ext = share_ext if is_share else static_ext
            else:
                ext = exec_ext

            package = target.package
            label_path = self.convert_package_name(package, target.name)

            output_dir = "$(LIB_DIR)/" if is_lib else "$(BIN_DIR)/"
            print(output_dir)
            package_build_path = self.convert_package_build_path(package, output_dir)
            package_bin_set.add(package_build_path)

            if not is_lib or is_share:
                BUILD_RULE = """\
{LABEL} : {DEPS} {PACKAGE_BUILD_PATH} {PACKAGE_BUILD_PATH}/{target}
{PACKAGE_BUILD_PATH}/{target}: {SRCS_OBJ}
\t$(LINK.cc) {COPTS} $^ -o $@ {INCLUDE} {LINKOPTS}
    """.format(
                    LABEL=label_path,
                    PACKAGE_BUILD_PATH=package_build_path,
                    target=target.name + ext,
                    DEPS=self.convert_deps(target.deps),
                    SRCS_OBJ=self.convert_src_obj(target.srcs, package),
                    INCLUDE=self.convert_opts(target.hdrs, "-I"),
                    COPTS=self.convert_opts(target.copts),
                    LINKOPTS=self.convert_opts(target.linkopts)
                )
                self.write_line(BUILD_RULE)
            else:
                BUILD_RULE = """\
{LABEL} : {DEPS} {PACKAGE_BUILD_PATH} {PACKAGE_BUILD_PATH}/{target}
{PACKAGE_BUILD_PATH}/{target}: {SRCS_OBJ}
\t$(AR) crv $@ $^
""".format(
                    LABEL=label_path,
                    PACKAGE_BUILD_PATH=package_build_path,
                    target=target.name + ext,
                    DEPS=self.convert_deps(target.deps),
                    SRCS_OBJ=self.convert_src_obj(target.srcs, package)
                )
                self.write_line(BUILD_RULE)
            target_list.append(label_path)

        for package_bin_path in package_bin_set:
            PACKAGE_DIR = "%s:\n\tmkdir -p %s\n" % (package_bin_path, package_bin_path)
            self.write_line(PACKAGE_DIR)

        ALL = """\
all: {TARGETS}
""".format(TARGETS=" ".join(target_list))
        self.write_line(ALL)

        END = """\
clean:
\trm -rf $(BIN_DIR) $(OBJ_DIR)

.PHONY:clean all
""".format(TARGET="")
        self.write_line(END)
        self.close()

    def build_target(self, label):
        target = self.convert_label(label)
        self.make_cmd(target)

    def make_cmd(self, target):
        cmd = "make -f {BUILD_DIR}/Makefile {TARGET}".format(BUILD_DIR=self.build_dir, TARGET=target)
        os.system(cmd)

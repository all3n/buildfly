import os

from buildfly.actions.base_action import BaseAction


class NewAction(BaseAction):
    CFG_FILE = "buildfly.yml"
    def parse_args(self, parser):
        parser.add_argument('name', metavar='name', type=str,
                            help="project name")
    TEMPLATE = """\
compiler: "cmake >= 3.10"
glibc: 2.12
main:
    type: bin
    srcs: ["src/*.cpp"]
    cflags: "-O3"
    libs: []
    includes: ["include"]
    deps: [
    ]
"""
    def run(self):
        project_dir = self.get_cur_file(self.args.name)
        if not os.path.exists(project_dir):
            os.makedirs(project_dir)
            cfg_new = self.get_cur_file(os.path.join(self.args.name, self.CFG_FILE))
            os.makedirs(os.path.join(project_dir, "src"))
            os.makedirs(os.path.join(project_dir, "include"))

            with open(cfg_new, 'w') as f:
                f.write(self.TEMPLATE)
        else:
            print(f"{project_dir} exists,skip ")




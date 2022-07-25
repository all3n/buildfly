import os

from buildfly.actions.base_action import BaseAction


class NewAction(BaseAction):
    CFG_FILE = "bfly_workspace.py"

    def parse_args(self, parser):
        parser.add_argument('-d', "--dir", default = "", type=str, help="project dir")

    TEMPLATE = """\
set_backend("cmake")
add_binary(
    "main",
    srcs = "src/**/*.cpp",
    includes = "include"
)
"""

    def run(self):
        project_dir = self.get_cur_file(self.args.dir)
        cfg_new = self.get_cur_file(os.path.join(self.args.dir, self.CFG_FILE))
        if not os.path.exists(cfg_new):
            if not os.path.exists(project_dir):
                os.makedirs(project_dir)

            os.makedirs(os.path.join(project_dir, "src"))
            os.makedirs(os.path.join(project_dir, "include"))
            with open(cfg_new, 'w') as f:
                f.write(self.TEMPLATE)
        else:
            print(f"{project_dir} exists,skip ")

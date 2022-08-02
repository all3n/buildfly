import os

from buildfly.actions.base_action import BaseAction
from buildfly.utils.system_utils import get_bfly_path,exec_cmd


class PkgAction(BaseAction):
    pkg_url = "https://github.com/all3n/buildfly-pkgs.git"

    def parse_args(self, parser):
        parser.add_argument('action', default = "update", type=str, help="project dir")

    def run(self):
        if self.args.action == 'update':
            pkgs_dir = get_bfly_path("pkgs")
            if os.path.exists(os.path.join(pkgs_dir, ".git")):
                cmd = f"cd {pkgs_dir};git pull"
                exec_cmd(cmd)
            else:
                cmd = f"git clone {self.pkg_url} {pkgs_dir}"
                exec_cmd(cmd)
        else:
            raise Exception("action not support")


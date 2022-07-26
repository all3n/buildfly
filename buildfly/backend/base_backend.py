import logging
import semver
import os
import tempfile

from buildfly.utils.github_api_utils import api_client
from buildfly.env import BENV
from buildfly.utils.system_utils import get_bfly_path, exec_cmd
from buildfly.utils.http_pkg_utils import download_http_pkg

logger = logging.getLogger(__name__)


class BaseBackend(object):
    def __init__(self, ctx):
        self.name = "base"
        self.ctx = ctx

    def setup(self):
        pass

    def generate(self):
        pass

    def build(self):
        pass

    def install_tool_if_required(self, name, github, tool_version_expression, env_bin=None, bin_ver_attr=None,
                                 pkg_os_pattern={}, bin_path={}):
        tool_bin = self.check_tool_exists(name, tool_version_expression, env_bin, bin_ver_attr, bin_path)
        if tool_bin:
            return tool_bin
        else:
            self.install_tools(name, github, tool_version_expression, pkg_os_pattern, bin_path)

    def check_tool_exists(self, name, tool_version_expression, env_bin=None, bin_ver=None, bin_path={}):
        tool_dir = get_bfly_path("tools/%s" % name)
        match = False
        match_tool = None
        if os.path.exists(tool_dir):
            tool_vers = os.listdir(tool_dir)
            cvs = sorted(list(map(semver.VersionInfo.parse, tool_vers)), reverse=True)

            if tool_version_expression == "latest":
                if cvs:
                    match = True
                    bpath = bin_path.get(BENV.system, None)
                    match_tool = os.path.join(tool_dir, str(cvs[0]), bpath)
            else:
                for cv in cvs:
                    if cv.match(tool_version_expression):
                        match = True
                        bpath = bin_path.get(BENV.system, None)
                        match_tool = os.path.join(tool_dir, str(cv), bpath)
                        break
        else:
            if env_bin and bin_ver:
                sv = semver.VersionInfo.parse(bin_ver)
                match = sv.match(tool_version_expression)
                if match:
                    match_tool = env_bin
                    logger.info(f"system {name} version {sv} [Y]")
                else:
                    logger.info(f"system {name} version {sv} [X]")
        if match:
            logger.info(f'match {name}: {match_tool}')
            return match_tool
        else:
            return None

    def install_tools(self, name, github, tool_version_expression, pkg_os_pattern={}, bin_path={}):
        owner, repo_name = github.split("/")
        releases = api_client.list_releases(owner, repo_name)
        tool_version = None
        tool_assets = None
        if tool_version_expression == "latest":
            tool_version, tool_assets = releases[0]
        else:
            for rv, assets in releases:
                sv = semver.VersionInfo.parse(rv.replace("v", ""))
                if sv.match(tool_version_expression):
                    tool_version = rv
                    tool_assets = assets
                    break
        if tool_version is None:
            return None
        logger.info(f"try install {tool_version}")
        os_pattern = pkg_os_pattern.get(BENV.system)

        asset_urls = [ast["browser_download_url"] for ast in tool_assets if os_pattern in ast["name"]]
        logger.info(f"{asset_urls}")
        if asset_urls:
            tool_version_dir = get_bfly_path("tools/%s/%s/" % (name, str(tool_version.replace("v", ""))))
            if not os.path.exists(tool_version_dir):
                os.makedirs(tool_version_dir)

            # use first
            # TODO
            asset_url = asset_urls[0]

            if asset_url.endswith(".tar.gz") or asset_url.endswith(".tgz"):
                suffix = ".tar.gz"
            elif asset_url.endswith(".zip"):
                suffix = ".zip"

            tmp_file = tempfile.NamedTemporaryFile(prefix=tool_version, suffix=suffix)
            logger.info("Download Release %s " % asset_url)
            tmp_file_path = tmp_file.name
            download_http_pkg(asset_url, tmp_file_path)
            if suffix == ".tar.gz":
                cmd = f"tar --strip-components=1 -zxvf {tmp_file_path} -C {tool_version_dir}"
            elif suffix == ".zip":
                cmd = f"unzip {tmp_file_path} -d {tool_version_dir}"
            exec_cmd(cmd)
            bpath = bin_path.get(BENV.system, None)
            tool_path = os.path.join(tool_version_dir, bpath)
            return tool_path
        return None

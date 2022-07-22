#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""
github api
    https://developer.github.com/v3/repos/#list-tags
"""
import json

import requests

from buildfly.config.global_config import G_CONFIG

GITHUB_API_V3 = "https://api.github.com"
proxy = G_CONFIG.get_value("proxy")


class github_api(object):
    def api_request(self, url):
        return requests.get(url, proxies=proxy)

    # GET /repos/:owner/:repo/tags
    def list_tags(self, owner, repo):
        api = "%s/repos/%s/%s/tags" % (GITHUB_API_V3, owner, repo)
        res = self.api_request(api)
        if res.status_code == 200:
            repo_tag_info = json.loads(res.text)
            t = {rti["name"]: rti for rti in repo_tag_info}
            return t
        return None

    def get_branch_info(self, owner, repo, branch):
        api = "%s/repos/%s/%s/branches/%s" % (GITHUB_API_V3, owner, repo, branch)
        res = self.api_request(api)
        if res.status_code == 200:
            repo_branch_info = json.loads(res.text)
            return repo_branch_info
        else:
            return None

    def is_repo_exists(self, owner, repo):
        api = "%s/repos/%s/%s" % (GITHUB_API_V3, owner, repo)
        res = self.api_request(api)
        if res.status_code == 200:
            repo_info = json.loads(res.text)
            return "id" in repo_info
        else:
            return False


api_client = github_api()

if __name__ == '__main__':
    gapi = github_api()
    gapi.list_tags("protocolbuffers", "protobuf")

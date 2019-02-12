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
import requests
import json

GITHUB_API_V3 = "https://api.github.com"

class github_api(object):
    # GET /repos/:owner/:repo/tags
    def list_tags(self, owner, repo):
        api = "%s/repos/%s/%s/tags" % (GITHUB_API_V3, owner, repo)
        res = requests.get(api)
        if res.status_code == 200:
            repo_tag_info = json.loads(res.text)
            t = {rti["name"] : {"tarball_url": rti["tarball_url"]} for rti in repo_tag_info}
            return t
        return None

api_client = github_api()



if __name__ == '__main__':
    gapi = github_api()
    gapi.list_tags("protocolbuffers", "protobuf")





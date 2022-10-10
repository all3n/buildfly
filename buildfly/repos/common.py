#!/usr/bin/env python
# -*- coding: utf-8 -*-
# Author wanghuacheng
# Description
# Date 10/10/22 8:20 PM
# Version 1.0

class BFlyPkg(object):
    def __init__(self, name, group, version):
        self.name = name
        self.group = group
        self.version = version
        self.kwargs = {}
        self.commit_sha = None

    def __repr__(self):
        return json.dumps(self.__dict__, indent=2)

    def path(self):
        p = []
        p.append(self.type)
        if self.group:
            p.append(self.group)
        p.append(self.name)
        if self.version:
            p.append("v")
            p.append(self.version)
        if self.commit_sha:
            p.append('commit')
            p.append(self.commit_sha)
        return "/".join(p)


class PkgType(enum.Enum):
    GITHUB = 1

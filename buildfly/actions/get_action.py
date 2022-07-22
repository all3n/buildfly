#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""

"""
from buildfly.actions.base_action import BaseAction
from buildfly.utils.dep_utils import *
from buildfly.utils.yaml_conf_utils import BuildDependency
import os


class GetAction(BaseAction):
    def parse_args(self, parser):
        parser.add_argument('libdesc', metavar='libdesc', type=str,
                            help="lib description")

    def run(self):
        libdesc = self.args.libdesc
        name = os.path.basename(libdesc).split("@")[0]
        app_dep = BuildDependency(name, libdesc)
        get_dep(app_dep)

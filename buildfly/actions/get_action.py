#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""

"""
from buildfly.actions.basic_action import basic_action
from buildfly.utils.dep_utils import *
class get_action(basic_action):
    def parse_args(self, parser):
        parser.add_argument('libdesc', metavar='libdesc', type=str,
                    help="lib description")

    def run(self):
        get_dep(self.args.libdesc)



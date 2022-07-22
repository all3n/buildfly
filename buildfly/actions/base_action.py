#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""
Action Base Class
"""
import os
import sys


class BaseAction(object):
    args = None

    def get_cur_dir(self):
        cur_dir = os.path.abspath(sys.path[0])
        return cur_dir

    def get_cur_file(self, f):
        return os.path.join(self.get_cur_dir(), f)

    def run(self):
        pass

    def parse_args(self, parser):
        pass

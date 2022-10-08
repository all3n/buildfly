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

from buildfly.utils.log_utils import get_logger

logger = get_logger(__name__)


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

    def split_args(self):
        args = [arg for arg in self.args.args if '=' not in arg]
        kwargs = dict([tuple(arg.split("=")) for arg in self.args.args if '=' in arg])
        return args, kwargs


class SubCmdAction(BaseAction):
    def parse_args(self, parser):
        parser.add_argument('cmd', metavar='cmd', type=str, nargs=1,
                            help="cmd")
        parser.add_argument('args', metavar='args', type=str, nargs="*",
                            help="cmd args")

    def run(self):
        cmd = self.args.cmd[0]
        args, kwargs = self.split_args()
        cmd_method = f"cmd_{cmd}"
        if hasattr(self, cmd_method):
            if args or kwargs:
                getattr(self, cmd_method)(*args, **kwargs)
            else:
                getattr(self, cmd_method)()
        else:
            logger.error(f"{cmd} not found")

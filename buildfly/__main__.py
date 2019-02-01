#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""

"""
from buildfly.args import *
import importlib
ACTION_PACKAGE="buildfly.actions"


def run_action(action):
    action_module = "%s.%s_action" % (ACTION_PACKAGE, action)
    action_class_name = "%s_action" % (action)
    action_class = getattr(importlib.import_module(action_module),action_class_name)
    action_obj = action_class()
    action_obj.run()


def main():
    action = ARGS.action
    run_action(action)



if __name__ == '__main__':
    main()

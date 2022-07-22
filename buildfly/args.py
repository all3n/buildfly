#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""
build cli args
"""
import argparse
import importlib
import os
import sys

ACTION_PACKAGE = "buildfly.actions"

actions_module = importlib.import_module(ACTION_PACKAGE)
action_module_abs_path = actions_module.__path__[0]
all_actions = [f.split("_")[0] for f in os.listdir(action_module_abs_path) if "_action" in f and f != "basic_action.py"]

parser = argparse.ArgumentParser(prog="bfly", description='buildfly action [options]')
parser.add_argument('action', metavar='action', type=str,
                    default="build",
                    help="actions: \n %s" % (all_actions))

app_args = sys.argv
if len(app_args) > 1:
    action = app_args[1]
else:
    print("need action")
    parser.print_help()
    sys.exit(-1)


def _build_action(action):
    if action not in all_actions:
        print("%s action is not valid" % (action))
        parser.print_help()
        sys.exit(-1)
    action_module = "%s.%s_action" % (ACTION_PACKAGE, action)
    action_class_name = "%sAction" % (action)
    action_class = getattr(importlib.import_module(action_module), action_class_name)
    action_obj = action_class()
    return action_obj


ACTION = _build_action(action)
ACTION.parse_args(parser)
ACTION.args = parser.parse_args()

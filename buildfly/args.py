#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""

"""
import argparse
import sys
import importlib
import os

ACTION_PACKAGE="buildfly.actions"

actions_module = importlib.import_module(ACTION_PACKAGE)
action_module_abs_path = actions_module.__path__._path[0]
all_actions = [f.split("_")[0] for f in os.listdir(action_module_abs_path) if "_action" in f and f != "basic_action.py"]

parser = argparse.ArgumentParser(prog="bfly",description='buildfly action [options]')
parser.add_argument('action', metavar='action', type=str,
                    default="build",
                    help="actions: \n %s" % (all_actions))

ARGS = parser.parse_args()

def get_action(action):
    if action not in all_actions:
        print("%s action is not valid" % (action))
        parser.print_help()
        sys.exit(-1)
    action_module = "%s.%s_action" % (ACTION_PACKAGE, action)
    action_class_name = "%s_action" % (action)
    action_class = getattr(importlib.import_module(action_module),action_class_name)
    action_obj = action_class()
    return action_obj

ACTION=get_action(ARGS.action)

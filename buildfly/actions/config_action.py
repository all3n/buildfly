#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""

"""
import sys
import os
import logging
import yaml

from buildfly.actions.basic_action import basic_action
from buildfly.utils.color_utils import *
from buildfly.utils.system_utils import get_bfly_path

CONFIG_FILE = get_bfly_path("bfly-config.yaml")

class config_action(basic_action):
    def parse_args(self, parser):
        parser.add_argument('name', metavar='name', type=str, nargs = 1,
                    help="name")

        parser.add_argument('value', metavar='value', type=str, nargs = "?",
                    help="value")
    def run(self):
        name = self.args.name[0]
        value = self.args.value
        sys_config = None
        if os.path.exists(CONFIG_FILE):
            with open(CONFIG_FILE, "r") as f:
                sys_config = yaml.load(f)
        if value:
            if not sys_config:
                sys_config = {}
            # set value
            name_fields = name.split(".")

            fv = sys_config
            for nf in name_fields[:-1]:
                if nf in fv:
                    fv = fv[nf]
                else:
                    fv[nf] = {}
                    fv = fv[nf]

            last_field = name_fields[-1]
            if type(fv) != dict:
                print("%s type must be dict: current is %s" % (name, type(fv)))
                sys.exit(-1)
            fv[last_field] = value
            with open(CONFIG_FILE, "w") as f:
                yaml.dump(sys_config, f)
            print("Saved %s : %s" % (name , value))
        else:
            # get value
            name_fields = name.split(".")
            fv = sys_config
            for nf in name_fields[:-1]:
                if nf in fv:
                    fv = fv[nf]
                else:
                    print("config key: %s not exists" % name)
                    sys.exit(-1)

            last_field = name_fields[-1]
            if last_field in fv:
                print("%s : %s" % (name , fv[last_field]))
            else:
                print("config key: %s not exists" % name)
                sys.exit(-1)

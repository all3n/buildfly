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
from buildfly.actions.basic_action import basic_action
from buildfly.utils.yaml_conf_utils import yaml_conf_loader
CONF_NAME="buildfly.yaml"
class build_action(basic_action):
    def run(self):
        cur_dir = os.path.abspath(sys.path[0])
        print("run build action in %s" % (cur_dir))
        self.parse_build_conf(os.path.join(cur_dir, CONF_NAME))

    def parse_build_conf(self, conf_file):
        if not os.path.exists(conf_file):
            print("%s not exist!" % conf_file)
            sys.exit(-1)

        app_conf = yaml_conf_loader(conf_file)
        print(app_conf)


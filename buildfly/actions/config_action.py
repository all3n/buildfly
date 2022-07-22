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
from buildfly.config.global_config import G_CONFIG


class ConfigAction(BaseAction):
    def parse_args(self, parser):
        parser.add_argument('name', metavar='name', type=str, nargs=1,
                            help="name")

        parser.add_argument('value', metavar='value', type=str, nargs="?",
                            help="value")

    def run(self):
        name = self.args.name[0]
        value = self.args.value

        if value:
            G_CONFIG.set_value(name, value, save=True)
        else:
            value = G_CONFIG.get_value(name)
            print("%s=%s" % (name, value))

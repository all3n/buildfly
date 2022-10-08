#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghch <wanghch@wanghch-MacBookPro.local>
#
# Distributed under terms of the MIT license.

"""

"""
import os
import yaml
import sys
from buildfly.utils.system_utils import get_bfly_path


class GlobalConfig(object):
    sys_config = {}

    def __init__(self, f):
        self.f = f
        if os.path.exists(f):
            with open(f, "r") as cf:
                self.sys_config = yaml.load(cf, yaml.FullLoader)
                if not self.sys_config:
                    self.sys_config = {}

    def set_value(self, name, value, save=False, entire=False):
        # set value
        name_fields = name.split(".")

        fv = self.sys_config
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
        if '[' in value or '{' in value:
            value = eval(value)
        else:
            value = str(value)
        if not entire and type(value) == dict:
            print("partial update %s" % name)
            if last_field in fv:
                fv[last_field].update_repo(value)
            else:
                fv[last_field] = value
        else:
            print("update %s" % last_field)
            fv[last_field] = value

        print("set %s : %s" % (name, value))
        if save:
            self.save()

    def get_value(self, name):
        name_fields = name.split(".")
        fv = self.sys_config
        for nf in name_fields[:-1]:
            if nf in fv:
                fv = fv[nf]
            else:
                return None

        last_field = name_fields[-1]
        if last_field in fv:
            return fv[last_field]
        else:
            return None

    def save(self):
        with open(self.f, "w") as f:
            yaml.dump(self.sys_config, f)


G_CONFIG = GlobalConfig(get_bfly_path("bfly-config.yaml"))

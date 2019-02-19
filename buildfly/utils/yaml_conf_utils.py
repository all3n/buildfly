#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""

"""
import yaml
from collections import namedtuple
BuildConf = namedtuple('BuildConf', ['bins','libs', 'args'])


def yaml_conf_loader(yaml_file):
    with open(yaml_file, "r", encoding='utf-8') as yf:
        conf =  yaml.load(yf.read())
        bins = {}
        libs = {}
        args = {}
        for k,v in conf.items():
            if type(v) == dict and 'type' in v:
                if v['type'] == 'bin':
                    bins[k] = v
                elif v['type'] == 'lib':
                    libs[k] = v
            else:
                args[k] = v

        build_conf = BuildConf(bins = bins, libs = libs, args = args)
        return build_conf

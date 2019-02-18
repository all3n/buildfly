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
def yaml_conf_loader(yaml_file):
    with open(yaml_file, "r", encoding='utf-8') as yf:
        return yaml.load(yf.read())

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

parser = argparse.ArgumentParser(description='buildfly action [options]')
parser.add_argument('action', metavar='action', type=str,
                    default="build",
                    help='process action')

ARGS = parser.parse_args()

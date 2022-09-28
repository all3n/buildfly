#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""

"""
import glob
import os
import re
import stat
import sys

from buildfly.actions.base_action import BaseAction
from buildfly.utils.color_utils import *
from buildfly.utils.dep_utils import get_dep, get_glibc, get_dep_compile_options, check_if_needed
from buildfly.utils.system_utils import *
from buildfly.backend import *
from buildfly.utils.yaml_conf_utils import yaml_conf_loader, BuildDependency
from buildfly.utils.api_utils import BuildFlyAPI, bfly_api_method
from buildfly.common import BFlyRepo, BFlyBin, BFlyLibrary, BFlyDep
from buildfly.utils.string_utils import camelize

import logging

logger = logging.getLogger(__name__)


class RunAction(BaseAction):

    def parse_args(self, parser):
        parser.add_argument('target', metavar='target', type=str, nargs="?",
                            help="build target")

    def run(self):
        pass

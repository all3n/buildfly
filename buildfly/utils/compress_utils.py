#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""

"""

import os
import tarfile


def members(tar, strip=1):
    for member in tar.getmembers():
        if strip > 0:
            member.path = member.path.split('/', strip)[-1]
        yield member


def compress_tar_gz(dir_need_compress, out_file):
    with tarfile.open(out_file, "w:gz") as tf:
        for root, d, files in os.walk(dir_need_compress):
            for f in files:
                full_path = os.path.join(root, f)
                tf.add(full_path)


def uncompress_tar_gz(dir_need_uncompress, input_file, strip=0):
    print(dir_need_uncompress)
    if not os.path.exists(dir_need_uncompress):
        os.makedirs(dir_need_uncompress)

    t = tarfile.open(input_file)
    t.extractall(dir_need_uncompress, members=members(t, strip))

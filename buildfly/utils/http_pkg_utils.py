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

import requests

from buildfly.config.global_config import G_CONFIG
from buildfly.utils.compress_utils import *
import logging
logger = logging.getLogger(__name__)


def format_size(bytes):
    try:
        bytes = float(bytes)
        kb = bytes / 1024
    except:
        print("传入的字节格式不对")
        return "Error"

    if kb >= 1024:
        M = kb / 1024
        if M >= 1024:
            G = M / 1024
            return "%.4fG" % (G)
        else:
            return "%.2fM" % (M)
    else:
        return "%.2fkb" % (kb)


def show_progress(i, content_length, fmt="{SPEED}"):
    sys.stdout.write((len(fmt) + 10) * " " + "\r")
    sys.stdout.flush()
    size_format = format_size(i)
    if content_length > 0:
        all_size_format = format_size(content_length)
        speed = "%s/%s" % (size_format, all_size_format)
    else:
        speed = "%s" % (size_format)
    log = fmt.replace("{SPEED}", speed)
    sys.stdout.write(log + "\r")
    sys.stdout.flush()


def download_http_pkg(url, tmp_pkg_file):
    pkg_base_dir = os.path.dirname(tmp_pkg_file)
    if not os.path.exists(pkg_base_dir):
        os.makedirs(pkg_base_dir)

    github_mirror = G_CONFIG.get_value("github.mirror")

    logger.info(f"download {url}")
    if github_mirror:
        url = url.replace("github.com", github_mirror)

    proxy = G_CONFIG.get_value("proxy")
    res = requests.get(url, stream=True, headers={'Accept-Encoding': None}, proxies=proxy)
    try:
        # 'Transfer-Encoding': 'chunked' chunked 类型 没有content-length
        if "Transfer-Encoding" in res.headers and res.headers["Transfer-Encoding"] == "chunked":
            content_length = 0
        else:
            content_length = int(res.headers['content-length'])

        res.raise_for_status()
        with open(tmp_pkg_file, 'wb') as tpf:
            write_data_size = 0
            for chunk in res.iter_content(100000):
                write_data_size += len(chunk)
                show_progress(write_data_size, content_length, fmt="DOWNLOAD:%s {SPEED}" % (url))
                tpf.write(chunk)
            print("%s download finished" % (url))


    except Exception as exc:
        print(exc)


if __name__ == '__main__':
    import os

    tmp_file = os.path.expanduser("~/.buildfly/tmp/protobuf.tar.gz")
    code_dir = os.path.expanduser("~/.buildfly/src/github.com/protocolbuffers/protobuf/3.7.0rc1")
    url = "https://codeload.github.com/google/googletest/tar.gz/release-1.8.1"
    # url = "https://codeload.github.com/protocolbuffers/protobuf/legacy.tar.gz/v3.7.0rc1"
    download_http_pkg(url, tmp_file)
    uncompress_tar_gz(code_dir, tmp_file)

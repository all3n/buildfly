#! /usr/bin/env python
# -*- coding: utf-8 -*-
# vim:fenc=utf-8
#
# Copyright © 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.

"""

"""
import requests
import os
import sys
from buildfly.utils.compress_utils import *


def formatSize(bytes):
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

def show_progress(i, content_length):
    size_format = formatSize(i)
    if content_length > 0:
        all_size_format = formatSize(content_length)
        sys.stdout.write("%s/%s\r" % (size_format,all_size_format))
    else:
        sys.stdout.write("%-10s\r" % (size_format))
    sys.stdout.flush()



def donwload_http_pkg(url,tmp_pkg_file):
    pkg_base_dir = os.path.dirname(tmp_pkg_file)
    if not os.path.exists(pkg_base_dir):
        os.makedirs(pkg_base_dir)

    res = requests.get(url, stream = True, headers={'Accept-Encoding': None})
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
                show_progress(write_data_size,content_length)
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
    donwload_http_pkg(url,tmp_file)
    uncompress_tar_gz(code_dir,tmp_file)


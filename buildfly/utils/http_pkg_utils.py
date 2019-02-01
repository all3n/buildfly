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

def show_progress(i,full):
    print("\r %d/%d" % (i,full))



def donwload_http_pkg(url,tmp_pkg_file):
    pkg_base_dir = os.path.dirname(tmp_pkg_file)
    if not os.path.exists(pkg_base_dir):
        os.makedirs(pkg_base_dir)

    res = requests.get(url)
    try:
        print(res.headers)
        #content_size = res.headers['content-length']
        res.raise_for_status()
        with open(tmp_pkg_file, 'wb') as tpf:
            write_data_size = 0
            for chunk in res.iter_content(100000):
                write_data_size += len(chunk)
        #        show_progress(write_data_size,content_size)
                tpf.write(chunk)


    except Exception as exc:
        print(exc)



if __name__ == '__main__':
    import os
    tmp_file = os.path.expanduser("~/.buildfly/tmp/protobuf.tar.gz")
    donwload_http_pkg("https://codeload.github.com/protocolbuffers/protobuf/legacy.tar.gz/v3.7.0rc1",tmp_file)

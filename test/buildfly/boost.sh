#! /bin/sh
#
# boost.sh
# Copyright (C) 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
#
# Distributed under terms of the MIT license.
#
# SRC_DIR
# LIB_VERSION

env

declare -A version_pkg_map=()
version_pkg_map["1.63.0"]="https://sourceforge.net/projects/boost/files/boost/1.63.0/boost_1_63_0.tar.gz"
version_pkg_map["1.69.0"]="https://sourceforge.net/projects/boost/files/boost/1.69.0/boost_1_69_0.tar.gz"

PKG_URL=$version_pkg_map[$LIB_VERSION]
if [[ -z $PKG_URL ]];then
    echo $LIB_VERSION not define
    exit -1
fi

download_pkg  $SRC_DIR

pushd $SRC_DIR

./bootstrap.sh --with-libraries=${INSTALL_MODULES} --prefix=${INSTALL_PREFIX}

./b2 -j32 variant=release define=_GLIBCXX_USE_CXX11_ABI=0 install

popd

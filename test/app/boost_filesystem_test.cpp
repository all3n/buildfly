/*
 * boost_filesystem_test.cpp
 * Copyright (C) 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
 *
 * Distributed under terms of the MIT license.
 */
#include <iostream>
#include <boost/filesystem.hpp>
using namespace boost::filesystem;

int main(int argc, char *argv[])
{
    if(argc < 2){
        std::cout<<"need path"<<std::endl;
        return 1;
    } 
    std::cout<<argv[1]<<" "<<file_size(argv[1])<<std::endl;
    return 0;
}



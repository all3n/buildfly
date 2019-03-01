/*
 * boost_program_option_test.cpp
 * Copyright (C) 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
 *
 * Distributed under terms of the MIT license.
 */
#include<iostream>
#include <boost/program_options.hpp>
using namespace std;
using namespace boost::program_options;
int main(int argc, char **argv)
{
    options_description desc("Allowed options");
    desc.add_options()
        ("help", "produce help message")
        ("compression", value<int>(), "set compression level")
    ;

    variables_map vm;
    store(parse_command_line(argc, argv, desc), vm);
    notify(vm);

    if (vm.count("help")) {
        cout << desc << "\n";
        return 1;
    }

    if (vm.count("compression")) {
        cout << "Compression level was set to " << vm["compression"].as<int>() << ".\n";
    } else {
        cout << "Compression level was not set.\n";
     }
}

/*
 * json.cpp
 * Copyright (C) 2019 wanghuacheng <wanghuacheng@wanghuacheng-PC>
 *
 * Distributed under terms of the MIT license.
 */

#include <json/json.h>
#include <iostream>

int main(int argc, char *argv[])
{
    const char* str = "{\"uploadid\": \"UP000000\",\"code\": 100,\"msg\": \"\",\"files\": \"\"}";

    Json::Reader reader;
    Json::Value root;
    if (reader.parse(str, root))  // reader将Json字符串解析到root，root将包含Json里所有子元素
    {
        std::string upload_id = root["uploadid"].asString();  // 访问节点，upload_id = "UP000000"
        int code = root["code"].asInt();    // 访问节点，code = 100
        std::cout<< code<<std::endl;
    }
    return 0;
}




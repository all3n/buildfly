cmake_minimum_required(VERSION 3.15)
project(venv-demo)

# 设置 C++ 标准
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 查找依赖
find_package(fmt REQUIRED)

# 创建可执行文件
add_executable(venv-demo main.cpp)

# 链接库
target_link_libraries(venv-demo PRIVATE fmt::fmt)

# 设置包含目录
target_include_directories(venv-demo PRIVATE ${CMAKE_CURRENT_SOURCE_DIR})

# 输出一些有用的信息
message(STATUS "Build type: ${CMAKE_BUILD_TYPE}")
message(STATUS "C++ compiler: ${CMAKE_CXX_COMPILER}")
message(STATUS "C++ compiler version: ${CMAKE_CXX_COMPILER_VERSION}")
message(STATUS "CMake version: ${CMAKE_VERSION}")
message(STATUS "Install prefix: ${CMAKE_INSTALL_PREFIX}")

# 如果在 BuildFly 环境中，显示环境信息
if(DEFINED ENV{BUILDFLY_ENV_ROOT})
    message(STATUS "BuildFly environment root: $ENV{BUILDFLY_ENV_ROOT}")
endif()

cmake_minimum_required(VERSION 3.16)
project(absl-example VERSION 1.0.0 LANGUAGES CXX)

# Set C++ standard
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Find Abseil (installed by buildfly)
find_package(absl REQUIRED)

# Create executable
add_executable(absl-demo main.cpp)

# Link Abseil libraries
target_link_libraries(absl-demo
    absl::strings
    absl::types
)

# Include directories
target_include_directories(absl-demo PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}
)

# Compiler-specific options
if(CMAKE_CXX_COMPILER_ID STREQUAL "GNU")
    target_compile_options(absl-demo PRIVATE -Wall -Wextra -O2)
elseif(CMAKE_CXX_COMPILER_ID STREQUAL "Clang")
    target_compile_options(absl-demo PRIVATE -Wall -Wextra -O2)
elseif(CMAKE_CXX_COMPILER_ID STREQUAL "MSVC")
    target_compile_options(absl-demo PRIVATE /W4 /O2)
endif()

# Add custom target for running the demo
add_custom_target(run-demo
    COMMAND absl-demo
    DEPENDS absl-demo
    WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
    COMMENT "Running Abseil demo program"
)
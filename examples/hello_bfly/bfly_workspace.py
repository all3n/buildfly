set("cmake_version", ">3.21.0")
set_backend("cmake")
add_binary(
    "main",
    srcs = "src/**/*.cpp",
    includes = "include"
)

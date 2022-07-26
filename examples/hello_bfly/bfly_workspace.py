
set("cmake_version", ">3.23.3")
set("compiler", "gcc")
set("compiler_version", ">5.5")
set("cmake_generator", "ninja")

set_backend("cmake")

add_binary(
    "main",
    srcs = ["src/**/*.cpp"],
    includes = ["include"]
)




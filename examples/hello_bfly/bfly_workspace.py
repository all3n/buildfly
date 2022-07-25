set("cmake_version", ">3.23.3")
set_backend("cmake")
add_binary(
    "main",
    srcs = ["src/**/*.cpp"],
    includes = ["include"]
)

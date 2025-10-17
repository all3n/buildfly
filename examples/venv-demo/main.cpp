#include <iostream>
#include <fmt/format.h>

int main() {
    std::string message = fmt::format("Hello from BuildFly C++ Virtual Environment!");
    std::cout << message << std::endl;
    
    // 显示环境信息
    std::cout << "\nEnvironment Information:" << std::endl;
    std::cout << "========================" << std::endl;
    
    // 检查编译器
    #ifdef __GNUC__
        std::cout << "Compiler: GCC " << __GNUC__ << "." << __GNUC_MINOR__ << "." << __GNUC_PATCHLEVEL__ << std::endl;
    #elif defined(__clang__)
        std::cout << "Compiler: Clang " << __clang_major__ << "." << __clang_minor__ << "." << __clang_patchlevel__ << std::endl;
    #elif defined(_MSC_VER)
        std::cout << "Compiler: MSVC " << _MSC_VER << std::endl;
    #else
        std::cout << "Compiler: Unknown" << std::endl;
    #endif
    
    // 检查 C++ 标准
    std::cout << "C++ Standard: ";
    #if __cplusplus == 202002L
        std::cout << "C++20" << std::endl;
    #elif __cplusplus == 201703L
        std::cout << "C++17" << std::endl;
    #elif __cplusplus == 201402L
        std::cout << "C++14" << std::endl;
    #elif __cplusplus == 201103L
        std::cout << "C++11" << std::endl;
    #else
        std::cout << "C++" << __cplusplus << std::endl;
    #endif
    
    return 0;
}

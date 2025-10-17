#include <iostream>
#include <string>
#include <vector>
#include <memory>

// Abseil headers
#include "absl/strings/str_cat.h"
#include "absl/strings/str_join.h"
#include "absl/strings/string_view.h"
#include "absl/types/optional.h"
#include "absl/types/span.h"

int main() {
    std::cout << "=== Abseil Demo Program ===" << std::endl;

    // Test absl::StrCat
    std::string hello = absl::StrCat("Hello", " ", "Abseil", "!");
    std::cout << "StrCat result: " << hello << std::endl;

    // Test absl::StrJoin
    std::vector<std::string> words = {"Optimized", "Dependency", "Management"};
    std::string joined = absl::StrJoin(words, " ");
    std::cout << "StrJoin result: " << joined << std::endl;

    // Test absl::string_view
    absl::string_view view(hello);
    std::cout << "String view: " << view.substr(0, 5) << std::endl;

    // Test absl::optional
    absl::optional<int> maybe_value = 42;
    if (maybe_value.has_value()) {
        std::cout << "Optional value: " << maybe_value.value() << std::endl;
    }

    // Test absl::Span
    std::vector<int> numbers = {1, 2, 3, 4, 5};
    absl::Span<int> span(numbers);
    std::cout << "Span elements: ";
    for (int num : span) {
        std::cout << num << " ";
    }
    std::cout << std::endl;

    std::cout << "=== Demo completed successfully! ===" << std::endl;
    return 0;
}
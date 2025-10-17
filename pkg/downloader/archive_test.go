package downloader

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"buildfly/pkg/config"
)

func TestArchiveDownloader_Verify(t *testing.T) {
	// 创建测试文件
	testContent := "Hello, World! This is a test file for checksum verification."

	// 计算正确的校验和
	md5Hash := md5.Sum([]byte(testContent))
	sha256Hash := sha256.Sum256([]byte(testContent))

	// 创建临时文件
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	downloader := &ArchiveDownloader{}

	tests := []struct {
		name       string
		dependency config.Dependency
		shouldPass bool
	}{
		{
			name: "Valid MD5",
			dependency: config.Dependency{
				Name: "test-md5",
				Source: config.SourceInfo{
					MD5: fmt.Sprintf("%x", md5Hash),
				},
			},
			shouldPass: true,
		},
		{
			name: "Invalid MD5",
			dependency: config.Dependency{
				Name: "test-invalid-md5",
				Source: config.SourceInfo{
					MD5: "invalid-md5-hash",
				},
			},
			shouldPass: false,
		},
		{
			name: "Valid SHA256",
			dependency: config.Dependency{
				Name: "test-sha256",
				Source: config.SourceInfo{
					SHA256: fmt.Sprintf("%x", sha256Hash),
				},
			},
			shouldPass: true,
		},
		{
			name: "Invalid SHA256",
			dependency: config.Dependency{
				Name: "test-invalid-sha256",
				Source: config.SourceInfo{
					SHA256: "invalid-sha256-hash",
				},
			},
			shouldPass: false,
		},
		{
			name: "No checksum",
			dependency: config.Dependency{
				Name:   "test-no-checksum",
				Source: config.SourceInfo{
					// 没有设置任何校验和
				},
			},
			shouldPass: true, // 没有校验和时应该通过
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := downloader.Verify(tt.dependency, testFile)

			if tt.shouldPass && err != nil {
				t.Errorf("Expected verification to pass, but got error: %v", err)
			}

			if !tt.shouldPass && err == nil {
				t.Errorf("Expected verification to fail, but it passed")
			}
		})
	}
}

func TestArchiveDownloader_GetChecksums(t *testing.T) {
	downloader := &ArchiveDownloader{}

	tests := []struct {
		name     string
		source   config.SourceInfo
		expected map[string]string
	}{
		{
			name: "MD5 only",
			source: config.SourceInfo{
				MD5: "test-md5-hash",
			},
			expected: map[string]string{
				"md5": "test-md5-hash",
			},
		},
		{
			name: "SHA256 only",
			source: config.SourceInfo{
				SHA256: "test-sha256-hash",
			},
			expected: map[string]string{
				"sha256": "test-sha256-hash",
			},
		},
		{
			name: "Multiple checksums",
			source: config.SourceInfo{
				MD5:    "test-md5-hash",
				SHA256: "test-sha256-hash",
			},
			expected: map[string]string{
				"md5":    "test-md5-hash",
				"sha256": "test-sha256-hash",
			},
		},
		{
			name: "Legacy hash field",
			source: config.SourceInfo{
				Hash: "test-legacy-hash",
			},
			expected: map[string]string{
				"sha256": "test-legacy-hash",
			},
		},
		{
			name: "Checksums map",
			source: config.SourceInfo{
				Checksums: map[string]string{
					"md5":    "test-checksums-md5",
					"sha256": "test-checksums-sha256",
				},
			},
			expected: map[string]string{
				"md5":    "test-checksums-md5",
				"sha256": "test-checksums-sha256",
			},
		},
		{
			name:   "No checksums",
			source: config.SourceInfo{
				// 没有任何校验和字段
			},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := downloader.getChecksums(tt.source)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d checksums, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if resultValue, exists := result[key]; !exists {
					t.Errorf("Expected checksum %s, but not found", key)
				} else if resultValue != expectedValue {
					t.Errorf("Expected %s=%s, got %s", key, expectedValue, resultValue)
				}
			}
		})
	}
}

func TestArchiveDownloader_VerifyChecksumWithAlgorithm(t *testing.T) {
	// 创建测试文件
	testContent := "Hello, World! This is a test file for checksum verification."

	// 计算正确的 MD5 哈希
	expectedMD5 := md5.Sum([]byte(testContent))

	// 创建临时文件
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	downloader := &ArchiveDownloader{}

	// 测试正确的 MD5 校验
	err := downloader.verifyChecksumWithAlgorithm(testFile, fmt.Sprintf("%x", expectedMD5), "md5")
	if err != nil {
		t.Errorf("Expected MD5 verification to pass, but got error: %v", err)
	}

	// 测试错误的 MD5 校验
	err = downloader.verifyChecksumWithAlgorithm(testFile, "wrong-hash", "md5")
	if err == nil {
		t.Error("Expected MD5 verification to fail with wrong hash, but it passed")
	}

	// 测试不支持的算法
	err = downloader.verifyChecksumWithAlgorithm(testFile, "some-hash", "unsupported")
	if err == nil {
		t.Error("Expected verification to fail with unsupported algorithm, but it passed")
	}
}

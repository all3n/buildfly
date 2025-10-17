package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "fileutils-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	srcFile := filepath.Join(tempDir, "test.txt")
	dstFile := filepath.Join(tempDir, "test_copy.txt")
	content := []byte("Hello, World!")

	// 写入源文件
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	// 测试复制文件
	if err := CopyFile(srcFile, dstFile); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// 验证文件是否存在
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Fatal("Destination file does not exist")
	}

	// 验证文件内容
	copiedContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(copiedContent) != string(content) {
		t.Errorf("Content mismatch: expected %s, got %s", string(content), string(copiedContent))
	}

	// 验证权限是否保留
	srcInfo, err := os.Stat(srcFile)
	if err != nil {
		t.Fatalf("Failed to stat source file: %v", err)
	}

	dstInfo, err := os.Stat(dstFile)
	if err != nil {
		t.Fatalf("Failed to stat destination file: %v", err)
	}

	if srcInfo.Mode() != dstInfo.Mode() {
		t.Errorf("Permission mismatch: expected %v, got %v", srcInfo.Mode(), dstInfo.Mode())
	}
}

func TestCopyDir(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "fileutils-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcDir := filepath.Join(tempDir, "src")
	dstDir := filepath.Join(tempDir, "dst")

	// 创建源目录结构
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("Failed to create source dir: %v", err)
	}

	// 创建子目录
	subDir := filepath.Join(srcDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// 创建文件
	files := map[string][]byte{
		"file1.txt":    []byte("Content 1"),
		"file2.txt":    []byte("Content 2"),
		"subdir/file3": []byte("Content 3"),
		"subdir/file4": []byte("Content 4"),
	}

	for relPath, content := range files {
		fullPath := filepath.Join(srcDir, relPath)
		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", fullPath, err)
		}
	}

	// 测试复制目录
	if err := CopyDir(srcDir, dstDir); err != nil {
		t.Fatalf("CopyDir failed: %v", err)
	}

	// 验证目录结构
	for relPath := range files {
		srcPath := filepath.Join(srcDir, relPath)
		dstPath := filepath.Join(dstDir, relPath)

		// 检查文件是否存在
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			t.Errorf("Destination file does not exist: %s", dstPath)
			continue
		}

		// 检查内容是否一致
		srcContent, err := os.ReadFile(srcPath)
		if err != nil {
			t.Fatalf("Failed to read source file %s: %v", srcPath, err)
		}

		dstContent, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatalf("Failed to read destination file %s: %v", dstPath, err)
		}

		if string(srcContent) != string(dstContent) {
			t.Errorf("Content mismatch for %s: expected %s, got %s",
				relPath, string(srcContent), string(dstContent))
		}

		// 检查权限是否保留
		srcInfo, err := os.Stat(srcPath)
		if err != nil {
			t.Fatalf("Failed to stat source file %s: %v", srcPath, err)
		}

		dstInfo, err := os.Stat(dstPath)
		if err != nil {
			t.Fatalf("Failed to stat destination file %s: %v", dstPath, err)
		}

		if srcInfo.Mode() != dstInfo.Mode() {
			t.Errorf("Permission mismatch for %s: expected %v, got %v",
				relPath, srcInfo.Mode(), dstInfo.Mode())
		}
	}
}

func TestCopyFileNonExistent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fileutils-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcFile := filepath.Join(tempDir, "nonexistent.txt")
	dstFile := filepath.Join(tempDir, "copy.txt")

	// 测试复制不存在的文件
	if err := CopyFile(srcFile, dstFile); err == nil {
		t.Error("Expected error when copying nonexistent file, got nil")
	}
}

func TestCopyDirNonExistent(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fileutils-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcDir := filepath.Join(tempDir, "nonexistent")
	dstDir := filepath.Join(tempDir, "copy")

	// 测试复制不存在的目录
	if err := CopyDir(srcDir, dstDir); err == nil {
		t.Error("Expected error when copying nonexistent directory, got nil")
	}
}

func TestCopyFileToNonExistentDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fileutils-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcFile := filepath.Join(tempDir, "test.txt")
	nonExistentDir := filepath.Join(tempDir, "nonexistent")
	dstFile := filepath.Join(nonExistentDir, "test.txt")

	// 创建源文件
	if err := os.WriteFile(srcFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// 测试复制到不存在的目录（应该自动创建）
	if err := CopyFile(srcFile, dstFile); err != nil {
		t.Errorf("CopyFile failed when destination directory doesn't exist: %v", err)
	}

	// 验证文件是否被创建
	if _, err := os.Stat(dstFile); os.IsNotExist(err) {
		t.Error("Destination file was not created")
	}
}

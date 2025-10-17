package downloader

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"buildfly/internal/errors"
	"buildfly/pkg/config"
)

// GitDownloader Git 下载器
type GitDownloader struct{}

// Download 从 Git 仓库下载
func (gd *GitDownloader) Download(ctx context.Context, dep config.Dependency, targetDir string, callback ProgressCallback) error {
	// 检查 git 命令是否可用
	if _, err := exec.LookPath("git"); err != nil {
		return errors.DownloadErrorWithCause(err, "git command not found")
	}

	// 如果目标目录已存在，先删除
	if _, err := os.Stat(targetDir); err == nil {
		if err := os.RemoveAll(targetDir); err != nil {
			return errors.DownloadErrorWithCause(err, fmt.Sprintf("failed to remove existing directory %s", targetDir))
		}
	}

	// 获取所有可用的 URL 并尝试克隆
	urls := dep.Source.GetAllAvailableURLs()
	var lastErr error

	for i, url := range urls {
		fmt.Printf("Attempting to clone from URL %d/%d: %s\n", i+1, len(urls), url)

		// 克隆仓库
		cloneArgs := []string{"clone", url, targetDir}
		if dep.Source.Tag != "" {
			cloneArgs = append([]string{"clone", "--branch", dep.Source.Tag, "--depth", "1"}, cloneArgs[1:]...)
		}

		cmd := exec.CommandContext(ctx, "git", cloneArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			lastErr = errors.DownloadErrorWithCause(err, fmt.Sprintf("failed to clone repository %s", url))
			fmt.Printf("Failed to clone from %s: %v\n", url, err)
			continue
		}

		fmt.Printf("Successfully cloned from: %s\n", url)

		// 如果没有指定标签但指定了版本，尝试切换到对应的提交或标签
		if dep.Source.Tag == "" && dep.Version != "" {
			if err := gd.checkoutVersion(ctx, targetDir, dep.Version); err != nil {
				lastErr = errors.DownloadErrorWithCause(err, fmt.Sprintf("failed to checkout version %s", dep.Version))
				fmt.Printf("Failed to checkout version %s: %v\n", dep.Version, err)
				// 清理失败的克隆
				os.RemoveAll(targetDir)
				continue
			}
		}

		return nil
	}

	// 所有 URL 都失败了
	if lastErr != nil {
		return lastErr
	}
	return errors.DownloadError("no valid Git URLs found for cloning")
}

// checkoutVersion 切换到指定版本
func (gd *GitDownloader) checkoutVersion(ctx context.Context, repoDir, version string) error {
	// 先尝试获取远程标签
	cmd := exec.CommandContext(ctx, "git", "fetch", "--tags", "--depth", "1")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		// 如果获取标签失败，尝试获取所有分支
		cmd = exec.CommandContext(ctx, "git", "fetch", "--all")
		cmd.Dir = repoDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to fetch remote references: %w", err)
		}
	}

	// 尝试切换到标签
	cmd = exec.CommandContext(ctx, "git", "checkout", version)
	cmd.Dir = repoDir
	if err := cmd.Run(); err == nil {
		return nil
	}

	// 如果标签不存在，尝试切换到分支
	cmd = exec.CommandContext(ctx, "git", "checkout", version)
	cmd.Dir = repoDir
	if err := cmd.Run(); err == nil {
		return nil
	}

	// 如果分支也不存在，尝试切换到提交
	cmd = exec.CommandContext(ctx, "git", "checkout", version)
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout version/tag/branch %s: %w", version, err)
	}

	return nil
}

// Verify 验证 Git 仓库
func (gd *GitDownloader) Verify(dep config.Dependency, repoPath string) error {
	// 检查是否是 Git 仓库
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return errors.DownloadError(fmt.Sprintf("%s is not a git repository", repoPath))
	}

	// 获取当前提交信息
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return errors.DownloadErrorWithCause(err, "failed to get git commit hash")
	}

	currentCommit := strings.TrimSpace(string(output))

	// 如果指定了标签，验证当前是否在正确的标签上
	if dep.Source.Tag != "" {
		cmd = exec.Command("git", "describe", "--tags", "--exact-match")
		cmd.Dir = repoPath
		output, err := cmd.Output()
		if err != nil {
			return errors.DownloadError(fmt.Sprintf("not on tag %s", dep.Source.Tag))
		}

		currentTag := strings.TrimSpace(string(output))
		if currentTag != dep.Source.Tag {
			return errors.DownloadError(fmt.Sprintf("expected tag %s, got %s", dep.Source.Tag, currentTag))
		}
	}

	// 如果指定了哈希，验证提交是否匹配
	if dep.Source.Hash != "" {
		if currentCommit != dep.Source.Hash {
			return errors.DownloadError(fmt.Sprintf("expected commit %s, got %s", dep.Source.Hash, currentCommit))
		}
	}

	return nil
}

// GetRemoteURL 获取远程仓库 URL
func (gd *GitDownloader) GetRemoteURL(repoPath string) (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", errors.DownloadErrorWithCause(err, "failed to get remote URL")
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCurrentTag 获取当前标签
func (gd *GitDownloader) GetCurrentTag(repoPath string) (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--exact-match")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", errors.DownloadErrorWithCause(err, "not on any tag")
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCurrentBranch 获取当前分支
func (gd *GitDownloader) GetCurrentBranch(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", errors.DownloadErrorWithCause(err, "failed to get current branch")
	}

	branch := strings.TrimSpace(string(output))
	if branch == "HEAD" {
		return "", errors.DownloadError("not on any branch (detached HEAD)")
	}

	return branch, nil
}

// GetCommitHash 获取当前提交哈希
func (gd *GitDownloader) GetCommitHash(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", errors.DownloadErrorWithCause(err, "failed to get commit hash")
	}

	return strings.TrimSpace(string(output)), nil
}

// IsCleanWorkingTree 检查工作树是否干净
func (gd *GitDownloader) IsCleanWorkingTree(repoPath string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return false, errors.DownloadErrorWithCause(err, "failed to check git status")
	}

	return len(strings.TrimSpace(string(output))) == 0, nil
}

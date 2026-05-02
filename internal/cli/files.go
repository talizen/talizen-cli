package cli

import (
	"bysir/talizen-cli/internal/talizen"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func remotePathToLocal(root string, remotePath string) (string, error) {
	remotePath = strings.TrimSpace(remotePath)
	if remotePath == "" || remotePath == "/" {
		return "", fmt.Errorf("invalid remote path: %q", remotePath)
	}

	clean := filepath.Clean(strings.TrimPrefix(remotePath, "/"))
	if clean == "." || strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("unsafe remote path: %s", remotePath)
	}

	return filepath.Join(root, clean), nil
}

func localPathToRemote(root string, localPath string) (string, error) {
	rel, err := filepath.Rel(root, localPath)
	if err != nil {
		return "", fmt.Errorf("relative path: %w", err)
	}
	if rel == "." || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path is outside sync dir: %s", localPath)
	}

	return "/" + filepath.ToSlash(rel), nil
}

func writeRemoteFiles(root string, files []talizen.File) error {
	err := os.MkdirAll(root, 0o755)
	if err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	for _, file := range files {
		if file.IsDir {
			continue
		}

		localPath, err := remotePathToLocal(root, file.Path)
		if err != nil {
			return err
		}

		err = os.MkdirAll(filepath.Dir(localPath), 0o755)
		if err != nil {
			return fmt.Errorf("create parent dir for %s: %w", file.Path, err)
		}

		err = os.WriteFile(localPath, []byte(file.Body), 0o644)
		if err != nil {
			return fmt.Errorf("write %s: %w", file.Path, err)
		}
	}

	return nil
}

func shouldSkipPath(path string) bool {
	base := filepath.Base(path)
	switch base {
	case ".git", ".talizen", "node_modules", "dist", "build":
		return true
	default:
		return false
	}
}

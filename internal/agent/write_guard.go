package agent

import (
	"errors"
	"path/filepath"
	"strings"
)

var ErrPathNotAllowed = errors.New("target path is outside the docs/content allowlist")

func ValidateTargetPath(repoPath, targetPath string) (string, error) {
	clean := filepath.Clean(strings.TrimSpace(targetPath))
	if clean == "." || clean == "" {
		return "", ErrPathNotAllowed
	}
	normalized := filepath.ToSlash(clean)
	if strings.HasPrefix(normalized, "docs/") || strings.HasPrefix(normalized, "content/") {
		return filepath.Join(repoPath, clean), nil
	}
	return "", ErrPathNotAllowed
}

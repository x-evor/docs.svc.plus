package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func EnsureFrontmatter(content, title string) string {
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "---") {
		return content
	}
	safeTitle := strings.TrimSpace(title)
	if safeTitle == "" {
		safeTitle = "Untitled"
	}
	return fmt.Sprintf("---\ntitle: %s\n---\n\n%s\n", safeTitle, strings.TrimSpace(content))
}

func WriteFile(absolutePath, content string) (int, error) {
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return 0, err
	}
	data := []byte(content)
	if err := os.WriteFile(absolutePath, data, 0o644); err != nil {
		return 0, err
	}
	return len(data), nil
}

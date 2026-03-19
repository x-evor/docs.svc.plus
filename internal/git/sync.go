package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func Pull(repoPath string) (bool, string, error) {
	cmd := exec.Command("git", "-C", repoPath, "pull", "--ff-only")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return false, strings.TrimSpace(stderr.String()), fmt.Errorf("git pull failed: %w", err)
	}
	output := strings.TrimSpace(stdout.String())
	if output == "" {
		output = strings.TrimSpace(stderr.String())
	}
	return true, output, nil
}

package gitutil

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(out.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

func Available() error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git binary not found in PATH")
	}
	return nil
}

func CloneOrPull(repo, dir string) (string, error) {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		if _, err := run(dir, "pull", "--ff-only"); err != nil {
			return "", err
		}
		return "updated", nil
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return "", err
	}
	if _, err := run("", "clone", "--depth", "1", "--quiet", repo, dir); err != nil {
		return "", err
	}
	return "cloned", nil
}

func LastCommit(repoDir, relPath string) (string, string, error) {
	out, err := run(repoDir, "log", "-1", "--format=%H|%cI", "--", relPath)
	if err != nil {
		return "", "", err
	}
	hash, date, _ := strings.Cut(out, "|")
	return hash, date, nil
}

package installer

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	TargetClaudeCode = "claude-code"
	TargetCodex      = "codex"
	TargetWeb        = "web"

	ScopePersonal = "personal"
	ScopeProject  = "project"
)

func ValidTarget(target string) bool {
	return target == TargetClaudeCode || target == TargetCodex || target == TargetWeb
}

func Destination(target, scope, name string) (string, error) {
	switch target {
	case TargetClaudeCode:
		if scope == ScopeProject {
			return filepath.Join(".claude", "skills", name), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".claude", "skills", name), nil
	case TargetCodex:
		if scope == ScopeProject {
			return filepath.Join(".codex", "skills", name), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".codex", "skills", name), nil
	case TargetWeb:
		return filepath.Join("skai-exports", name+".zip"), nil
	}
	return "", fmt.Errorf("unknown target %q, valid targets are claude-code, codex, web", target)
}

func CopyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, info.Mode().Perm())
	})
}

func Symlink(src, dst string) error {
	abs, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.Symlink(abs, dst)
}

func ZipDir(src, name, zipPath string) error {
	if err := os.MkdirAll(filepath.Dir(zipPath), 0o755); err != nil {
		return err
	}
	out, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer out.Close()
	zw := zip.NewWriter(out)
	defer zw.Close()
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || d.Type()&fs.ModeSymlink != 0 {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		w, err := zw.Create(name + "/" + filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(w, f)
		return err
	})
}

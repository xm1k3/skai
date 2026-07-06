package scanner

import (
	"io/fs"
	"path/filepath"

	"github.com/xm1k3/skai/internal/config"
	"github.com/xm1k3/skai/internal/gitutil"
	"github.com/xm1k3/skai/internal/index"
	"github.com/xm1k3/skai/internal/skill"
)

var skippedDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
}

func ScanSource(src config.Source, dir string) []index.Skill {
	var skills []index.Skill
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && skippedDirs[d.Name()] {
			return filepath.SkipDir
		}
		if d.IsDir() || d.Name() != "SKILL.md" {
			return nil
		}
		skillDir := filepath.Dir(path)
		p, err := skill.Parse(skillDir)
		if err != nil || p.Name == "" || p.Description == "" {
			return nil
		}
		a := skill.Analyze(p)
		rel, err := filepath.Rel(dir, skillDir)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		hash, date, _ := gitutil.LastCommit(dir, rel)
		skills = append(skills, index.Skill{
			Name:                      p.Name,
			Description:               p.Description,
			Path:                      rel,
			Source:                    src.Name,
			SourceRepo:                src.Repo,
			LastCommitHash:            hash,
			LastCommitDate:            date,
			HasScripts:                a.HasScripts,
			NetworkCalls:              a.NetworkCalls,
			DestructiveOps:            a.DestructiveOps,
			ConfirmsBeforeDestructive: a.ConfirmsBeforeDestructive,
			ClaudeCodeOnly:            a.ClaudeCodeOnly,
			FrontmatterExtraFields:    a.ExtraFields,
			LineCount:                 a.LineCount,
			TokenEstimate:             a.TokenEstimate,
			Category:                  a.Category,
			ContentHash:               a.ContentHash,
		})
		return nil
	})
	return skills
}

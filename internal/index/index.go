package index

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xm1k3/skai/internal/config"
)

type Skill struct {
	Name                      string   `json:"name"`
	Description               string   `json:"description"`
	Path                      string   `json:"path"`
	Source                    string   `json:"source"`
	SourceRepo                string   `json:"source_repo"`
	LastCommitHash            string   `json:"last_commit_hash"`
	LastCommitDate            string   `json:"last_commit_date"`
	HasScripts                bool     `json:"has_scripts"`
	NetworkCalls              bool     `json:"network_calls"`
	DestructiveOps            bool     `json:"destructive_ops"`
	ConfirmsBeforeDestructive bool     `json:"confirms_before_destructive"`
	ClaudeCodeOnly            bool     `json:"claude_code_only"`
	FrontmatterExtraFields    []string `json:"frontmatter_extra_fields"`
	LineCount                 int      `json:"line_count"`
	TokenEstimate             int      `json:"token_estimate"`
	Category                  string   `json:"category"`
	ContentHash               string   `json:"content_hash"`
}

type Index struct {
	UpdatedAt time.Time `json:"updated_at"`
	Skills    []Skill   `json:"skills"`
}

func Load() (*Index, error) {
	path, err := config.IndexPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("index not found at %s, run skai sync first", path)
		}
		return nil, err
	}
	var idx Index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("corrupted index, run skai sync to rebuild: %w", err)
	}
	return &idx, nil
}

func Save(idx *Index) error {
	path, err := config.IndexPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (idx *Index) Find(query string) []Skill {
	var matches []Skill
	if source, name, ok := strings.Cut(query, "/"); ok {
		for _, s := range idx.Skills {
			if strings.EqualFold(s.Source, source) && strings.EqualFold(s.Name, name) {
				matches = append(matches, s)
			}
		}
		return matches
	}
	for _, s := range idx.Skills {
		if strings.EqualFold(s.Name, query) {
			matches = append(matches, s)
		}
	}
	return matches
}

func (s Skill) QualifiedName() string {
	return s.Source + "/" + s.Name
}

func (s Skill) RiskLevel() string {
	if s.DestructiveOps && !s.ConfirmsBeforeDestructive {
		return "high"
	}
	if s.DestructiveOps || s.NetworkCalls {
		return "medium"
	}
	return "low"
}

func (s Skill) LocalDir() (string, error) {
	sourcesDir, err := config.SourcesDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(sourcesDir, s.Source, filepath.FromSlash(s.Path)), nil
}

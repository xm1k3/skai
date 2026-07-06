package skill

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const maxScriptFileSize = 512 * 1024

var (
	networkRe     = regexp.MustCompile(`(?i)\bcurl\b|\bwget\b|\bfetch\(|https?://`)
	destructiveRe = regexp.MustCompile(`(?i)rm -rf|\brm\s+-\w|\brm\s+/|\bmv\s+\S|drop\s+table|\bdelete\b`)
	confirmRe     = regexp.MustCompile(`(?i)\bconfirm\b|confirmation|are you sure|ask before|ask the user|explicit approval`)
	scriptBlockRe = regexp.MustCompile("```(?:bash|sh|shell|zsh|python|py|javascript|js)\\b")
)

var claudeCodeOnlyFields = map[string]bool{
	"allowed-tools":            true,
	"context":                  true,
	"hooks":                    true,
	"disable-model-invocation": true,
}

var categories = []struct {
	name     string
	keywords []string
}{
	{"security", []string{"security", "vulnerab", "pentest", "exploit", "cve", "malware", "threat model", "audit"}},
	{"documents", []string{"pdf", "docx", "excel", "spreadsheet", "powerpoint", "slides", "document"}},
	{"devops", []string{"docker", "kubernetes", "terraform", "deploy", "ci/cd", "aws", "gcp", "azure", "infrastructure"}},
	{"data", []string{"sql", "database", "csv", "analytics", "dataset", "etl", "data pipeline", "visualization"}},
	{"testing", []string{"test", "coverage", "lint", "debugging"}},
	{"git", []string{"git", "github", "pull request", "commit", "changelog"}},
	{"web", []string{"web", "frontend", "html", "css", "react", "browser", "seo", "api"}},
	{"writing", []string{"writing", "blog", "article", "copywriting", "documentation"}},
	{"productivity", []string{"email", "calendar", "notes", "task", "workflow", "meeting", "planning"}},
	{"ai", []string{"llm", "prompt", "agent", "rag", "embedding", "model"}},
}

type Analysis struct {
	HasScripts                bool
	NetworkCalls              bool
	DestructiveOps            bool
	ConfirmsBeforeDestructive bool
	ClaudeCodeOnly            bool
	ExtraFields               []string
	LineCount                 int
	TokenEstimate             int
	Category                  string
	ContentHash               string
}

func Analyze(p *Parsed) Analysis {
	var a Analysis
	scripts := readScripts(p.Dir)
	combined := p.Body + "\n" + scripts
	if info, err := os.Stat(filepath.Join(p.Dir, "scripts")); err == nil && info.IsDir() {
		a.HasScripts = true
	} else if scriptBlockRe.MatchString(p.Body) {
		a.HasScripts = true
	}
	a.NetworkCalls = networkRe.MatchString(combined)
	a.DestructiveOps = destructiveRe.MatchString(combined)
	if a.DestructiveOps {
		a.ConfirmsBeforeDestructive = confirmRe.MatchString(combined)
	}
	for key := range p.Frontmatter {
		if key == "name" || key == "description" {
			continue
		}
		a.ExtraFields = append(a.ExtraFields, key)
		if claudeCodeOnlyFields[key] {
			a.ClaudeCodeOnly = true
		}
	}
	sort.Strings(a.ExtraFields)
	a.LineCount = strings.Count(p.Raw, "\n") + 1
	a.TokenEstimate = len(p.Raw) / 4
	a.Category = categorize(p.Name + " " + p.Description)
	sum := sha256.Sum256([]byte(p.Raw))
	a.ContentHash = hex.EncodeToString(sum[:])
	return a
}

func readScripts(dir string) string {
	scriptsDir := filepath.Join(dir, "scripts")
	var sb strings.Builder
	filepath.WalkDir(scriptsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil || info.Size() > maxScriptFileSize {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		sb.Write(data)
		sb.WriteString("\n")
		return nil
	})
	return sb.String()
}

func categorize(text string) string {
	lower := strings.ToLower(text)
	for _, c := range categories {
		for _, kw := range c.keywords {
			if strings.Contains(lower, kw) {
				return c.name
			}
		}
	}
	return "uncategorized"
}

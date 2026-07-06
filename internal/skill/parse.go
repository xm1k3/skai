package skill

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Parsed struct {
	Dir         string
	Name        string
	Description string
	Frontmatter map[string]any
	Body        string
	Raw         string
}

func Parse(dir string) (*Parsed, error) {
	data, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		return nil, err
	}
	raw := string(data)
	fm, body, err := splitFrontmatter(raw)
	if err != nil {
		return nil, err
	}
	meta := map[string]any{}
	if err := yaml.Unmarshal([]byte(fm), &meta); err != nil {
		return nil, errors.New("invalid frontmatter yaml: " + err.Error())
	}
	p := &Parsed{Dir: dir, Frontmatter: meta, Body: body, Raw: raw}
	if v, ok := meta["name"].(string); ok {
		p.Name = strings.TrimSpace(v)
	}
	if v, ok := meta["description"].(string); ok {
		p.Description = strings.TrimSpace(strings.ReplaceAll(v, "\n", " "))
	}
	return p, nil
}

func splitFrontmatter(content string) (string, string, error) {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return "", "", errors.New("SKILL.md does not start with yaml frontmatter")
	}
	rest := normalized[4:]
	if idx := strings.Index(rest, "\n---\n"); idx >= 0 {
		return rest[:idx], rest[idx+5:], nil
	}
	if strings.HasSuffix(rest, "\n---") {
		return rest[:len(rest)-4], "", nil
	}
	return "", "", errors.New("SKILL.md frontmatter is not terminated")
}

package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Severity string

const (
	SeverityError   Severity = "ERROR"
	SeverityWarning Severity = "WARN"
)

type Issue struct {
	Severity Severity
	Message  string
}

var (
	nameRe    = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	mdLinkRe  = regexp.MustCompile(`\]\(([^)\s]+)\)`)
	pathRefRe = regexp.MustCompile("`((?:scripts|references|assets)/[^`\\s]+)`")
)

func Validate(dir string) []Issue {
	var issues []Issue
	p, err := Parse(dir)
	if err != nil {
		return []Issue{{SeverityError, err.Error()}}
	}
	base := filepath.Base(dir)
	if p.Name == "" {
		issues = append(issues, Issue{SeverityError, "frontmatter field name is missing"})
	} else {
		if p.Name != base {
			issues = append(issues, Issue{SeverityError, fmt.Sprintf("frontmatter name %q does not match directory name %q", p.Name, base)})
		}
		if len(p.Name) > 64 {
			issues = append(issues, Issue{SeverityError, fmt.Sprintf("name is %d characters, maximum is 64", len(p.Name))})
		}
		if !nameRe.MatchString(p.Name) {
			issues = append(issues, Issue{SeverityWarning, "name should contain only lowercase letters, digits and hyphens"})
		}
	}
	if p.Description == "" {
		issues = append(issues, Issue{SeverityError, "frontmatter field description is missing"})
	} else if len(p.Description) > 1024 {
		issues = append(issues, Issue{SeverityError, fmt.Sprintf("description is %d characters, maximum is 1024", len(p.Description))})
	} else if len(p.Description) > 200 {
		issues = append(issues, Issue{SeverityWarning, fmt.Sprintf("description is %d characters, agents work best under 200", len(p.Description))})
	}
	lineCount := strings.Count(p.Raw, "\n") + 1
	if lineCount > 500 {
		issues = append(issues, Issue{SeverityWarning, fmt.Sprintf("SKILL.md has %d lines, consider moving detail into references/", lineCount)})
	}
	for _, ref := range referencedPaths(p.Body) {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(ref))); err != nil {
			issues = append(issues, Issue{SeverityWarning, fmt.Sprintf("referenced path %s does not exist", ref)})
		}
	}
	return issues
}

func referencedPaths(body string) []string {
	seen := map[string]bool{}
	var refs []string
	add := func(ref string) {
		ref = strings.TrimRight(ref, ".,;:!?")
		if fragment := strings.Index(ref, "#"); fragment >= 0 {
			ref = ref[:fragment]
		}
		if ref == "" || seen[ref] {
			return
		}
		seen[ref] = true
		refs = append(refs, ref)
	}
	for _, m := range mdLinkRe.FindAllStringSubmatch(body, -1) {
		ref := m[1]
		if strings.Contains(ref, "://") || strings.HasPrefix(ref, "#") ||
			strings.HasPrefix(ref, "mailto:") || strings.HasPrefix(ref, "/") ||
			strings.Contains(ref, "%") || strings.Contains(ref, "{") {
			continue
		}
		add(ref)
	}
	for _, m := range pathRefRe.FindAllStringSubmatch(body, -1) {
		if strings.Contains(m[1], "{") || strings.Contains(m[1], "<") || strings.Contains(m[1], "*") {
			continue
		}
		add(m[1])
	}
	return refs
}

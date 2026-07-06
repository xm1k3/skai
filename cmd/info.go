package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xm1k3/skai/internal/index"
	"github.com/xm1k3/skai/internal/installer"
)

var infoCmd = &cobra.Command{
	Use:   "info <skill>",
	Short: "Show full metadata and risk flags for a skill",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		idx, err := index.Load()
		if err != nil {
			return err
		}
		matches := idx.Find(args[0])
		if len(matches) == 0 {
			return fmt.Errorf("skill %q not found in index", args[0])
		}
		records, _ := installer.LoadManifest()
		for i, s := range matches {
			if i > 0 {
				fmt.Println()
			}
			printSkill(s, records)
		}
		if len(matches) > 1 {
			fmt.Printf("\n%d skills share the name %q, use source/name to disambiguate\n", len(matches), args[0])
		}
		return nil
	},
}

func printSkill(s index.Skill, records []installer.Record) {
	fmt.Printf("Name:                        %s\n", s.Name)
	fmt.Printf("Description:                 %s\n", s.Description)
	fmt.Printf("Category:                    %s\n", s.Category)
	fmt.Printf("Source:                      %s\n", s.Source)
	fmt.Printf("Repository:                  %s\n", s.SourceRepo)
	fmt.Printf("Path:                        %s\n", s.Path)
	fmt.Printf("Last commit:                 %s (%s)\n", shortHash(s.LastCommitHash), s.LastCommitDate)
	fmt.Printf("Risk level:                  %s\n", s.RiskLevel())
	fmt.Printf("Network calls:               %s\n", boolLabel(s.NetworkCalls))
	fmt.Printf("Destructive ops:             %s\n", boolLabel(s.DestructiveOps))
	fmt.Printf("Confirms before destructive: %s\n", boolLabel(s.ConfirmsBeforeDestructive))
	fmt.Printf("Has scripts:                 %s\n", boolLabel(s.HasScripts))
	fmt.Printf("Claude Code only:            %s\n", boolLabel(s.ClaudeCodeOnly))
	extra := "none"
	if len(s.FrontmatterExtraFields) > 0 {
		extra = strings.Join(s.FrontmatterExtraFields, ", ")
	}
	fmt.Printf("Extra frontmatter fields:    %s\n", extra)
	fmt.Printf("Lines:                       %d\n", s.LineCount)
	fmt.Printf("Token estimate:              %d\n", s.TokenEstimate)
	fmt.Printf("Content hash:                %s\n", shortHash(s.ContentHash))
	var installed []string
	for _, r := range records {
		if r.Name == s.Name && r.Source == s.Source {
			installed = append(installed, fmt.Sprintf("%s (%s) at %s", r.Target, r.Scope, r.Path))
		}
	}
	if len(installed) > 0 {
		fmt.Printf("Installed:                   %s\n", strings.Join(installed, "; "))
	}
}

func shortHash(hash string) string {
	if len(hash) > 12 {
		return hash[:12]
	}
	if hash == "" {
		return "unknown"
	}
	return hash
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

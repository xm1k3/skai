package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xm1k3/skai/internal/index"
)

var exportOutput string

var exportCmd = &cobra.Command{
	Use:   "export-awesome-list",
	Short: "Generate an awesome-list markdown file from the catalog",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		idx, err := index.Load()
		if err != nil {
			return err
		}
		byCategory := map[string][]index.Skill{}
		sources := map[string]bool{}
		for _, s := range idx.Skills {
			byCategory[s.Category] = append(byCategory[s.Category], s)
			sources[s.Source] = true
		}
		var categories []string
		for c := range byCategory {
			categories = append(categories, c)
		}
		sort.Strings(categories)
		var sb strings.Builder
		sb.WriteString("# Awesome Agent Skills\n\n")
		sb.WriteString(fmt.Sprintf("A curated list of %d Agent Skills aggregated from %d community sources.\n\n", len(idx.Skills), len(sources)))
		sb.WriteString("Generated with [skai](https://github.com/xm1k3/skai).\n\n")
		sb.WriteString("## Contents\n\n")
		for _, c := range categories {
			sb.WriteString(fmt.Sprintf("- [%s](#%s) (%d)\n", c, c, len(byCategory[c])))
		}
		sb.WriteString("\n")
		for _, c := range categories {
			sb.WriteString(fmt.Sprintf("## %s\n\n", c))
			skills := byCategory[c]
			sort.Slice(skills, func(i, j int) bool {
				return skills[i].Name < skills[j].Name
			})
			for _, s := range skills {
				desc := strings.ReplaceAll(s.Description, "\n", " ")
				sb.WriteString(fmt.Sprintf("- **%s** - %s ([%s](%s))\n", s.Name, desc, s.Source, s.SourceRepo))
			}
			sb.WriteString("\n")
		}
		if err := os.WriteFile(exportOutput, []byte(sb.String()), 0o644); err != nil {
			return err
		}
		fmt.Printf("Wrote %s with %d skills in %d categories\n", exportOutput, len(idx.Skills), len(categories))
		return nil
	},
}

func init() {
	exportCmd.Flags().StringVar(&exportOutput, "output", "awesome-list.md", "output file path")
	rootCmd.AddCommand(exportCmd)
}

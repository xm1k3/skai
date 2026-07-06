package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xm1k3/skai/internal/index"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Fuzzy search skills by name and description",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		idx, err := index.Load()
		if err != nil {
			return err
		}
		query := strings.ToLower(strings.Join(args, " "))
		type scored struct {
			skill index.Skill
			score int
		}
		var results []scored
		for _, s := range idx.Skills {
			score := scoreSkill(s, query)
			if score > 0 {
				results = append(results, scored{s, score})
			}
		}
		if len(results) == 0 {
			fmt.Printf("no skills matching %q\n", query)
			return nil
		}
		sort.Slice(results, func(i, j int) bool {
			if results[i].score != results[j].score {
				return results[i].score > results[j].score
			}
			return results[i].skill.Name < results[j].skill.Name
		})
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tCATEGORY\tRISK\tSOURCE\tDESCRIPTION")
		for _, r := range results {
			s := r.skill
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", s.Name, s.Category, s.RiskLevel(), s.Source, truncate(s.Description, 60))
		}
		w.Flush()
		fmt.Printf("\n%d skills\n", len(results))
		return nil
	},
}

func scoreSkill(s index.Skill, query string) int {
	name := strings.ToLower(s.Name)
	desc := strings.ToLower(s.Description)
	if name == query {
		return 1000
	}
	score := 0
	if strings.Contains(name, query) {
		score += 200
	}
	if strings.Contains(desc, query) {
		score += 100
	}
	for _, token := range strings.Fields(query) {
		if strings.Contains(name, token) {
			score += 40
		}
		if strings.Contains(desc, token) {
			score += 20
		}
	}
	if score == 0 && isSubsequence(query, name) {
		score = 10
	}
	return score
}

func isSubsequence(needle, haystack string) bool {
	i := 0
	for j := 0; j < len(haystack) && i < len(needle); j++ {
		if needle[i] == haystack[j] {
			i++
		}
	}
	return i == len(needle)
}

func init() {
	rootCmd.AddCommand(searchCmd)
}

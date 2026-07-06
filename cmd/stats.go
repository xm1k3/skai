package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xm1k3/skai/internal/index"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show catalog statistics by category, source and risk",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		idx, err := index.Load()
		if err != nil {
			return err
		}
		byCategory := map[string]int{}
		bySource := map[string]int{}
		byRisk := map[string]int{}
		scripts, network, destructive, claudeOnly := 0, 0, 0, 0
		for _, s := range idx.Skills {
			byCategory[s.Category]++
			bySource[s.Source]++
			byRisk[s.RiskLevel()]++
			if s.HasScripts {
				scripts++
			}
			if s.NetworkCalls {
				network++
			}
			if s.DestructiveOps {
				destructive++
			}
			if s.ClaudeCodeOnly {
				claudeOnly++
			}
		}
		fmt.Printf("Total skills: %d\n", len(idx.Skills))
		fmt.Printf("Last sync:    %s\n\n", idx.UpdatedAt.Format("2006-01-02 15:04:05 MST"))
		printCounts("BY CATEGORY", byCategory)
		printCounts("BY SOURCE", bySource)
		printCounts("BY RISK", byRisk)
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "FLAGS\tCOUNT")
		fmt.Fprintf(w, "has scripts\t%d\n", scripts)
		fmt.Fprintf(w, "network calls\t%d\n", network)
		fmt.Fprintf(w, "destructive ops\t%d\n", destructive)
		fmt.Fprintf(w, "claude code only\t%d\n", claudeOnly)
		w.Flush()
		return nil
	},
}

func printCounts(title string, counts map[string]int) {
	type row struct {
		key   string
		count int
	}
	var rows []row
	for k, v := range counts {
		rows = append(rows, row{k, v})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].count != rows[j].count {
			return rows[i].count > rows[j].count
		}
		return rows[i].key < rows[j].key
	})
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintf(w, "%s\tCOUNT\n", title)
	for _, r := range rows {
		fmt.Fprintf(w, "%s\t%d\n", r.key, r.count)
	}
	w.Flush()
	fmt.Println()
}

func init() {
	rootCmd.AddCommand(statsCmd)
}

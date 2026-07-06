package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xm1k3/skai/internal/index"
	"github.com/xm1k3/skai/internal/installer"
)

var (
	listCategory    string
	listRisk        string
	listTool        string
	listHasScripts  bool
	listNetwork     bool
	listDestructive bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List indexed skills with optional filters",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if listRisk != "" && listRisk != "low" && listRisk != "medium" && listRisk != "high" {
			return fmt.Errorf("invalid --risk %q, valid values are low, medium, high", listRisk)
		}
		if listTool != "" && !installer.ValidTarget(listTool) {
			return fmt.Errorf("invalid --tool %q, valid values are claude-code, codex, web", listTool)
		}
		idx, err := index.Load()
		if err != nil {
			return err
		}
		var filtered []index.Skill
		for _, s := range idx.Skills {
			if listCategory != "" && s.Category != listCategory {
				continue
			}
			if listRisk != "" && s.RiskLevel() != listRisk {
				continue
			}
			if (listTool == installer.TargetCodex || listTool == installer.TargetWeb) && s.ClaudeCodeOnly {
				continue
			}
			if listHasScripts && !s.HasScripts {
				continue
			}
			if listNetwork && !s.NetworkCalls {
				continue
			}
			if listDestructive && !s.DestructiveOps {
				continue
			}
			filtered = append(filtered, s)
		}
		sort.Slice(filtered, func(i, j int) bool {
			if filtered[i].Name != filtered[j].Name {
				return filtered[i].Name < filtered[j].Name
			}
			return filtered[i].Source < filtered[j].Source
		})
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tCATEGORY\tRISK\tSOURCE\tDESCRIPTION")
		for _, s := range filtered {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", s.Name, s.Category, s.RiskLevel(), s.Source, truncate(s.Description, 60))
		}
		w.Flush()
		fmt.Printf("\n%d skills\n", len(filtered))
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listCategory, "category", "", "filter by category")
	listCmd.Flags().StringVar(&listRisk, "risk", "", "filter by risk level (low, medium, high)")
	listCmd.Flags().StringVar(&listTool, "tool", "", "filter by compatible tool (claude-code, codex, web)")
	listCmd.Flags().BoolVar(&listHasScripts, "has-scripts", false, "only skills that ship executable scripts")
	listCmd.Flags().BoolVar(&listNetwork, "network", false, "only skills that perform network calls")
	listCmd.Flags().BoolVar(&listDestructive, "destructive", false, "only skills with destructive operations")
	rootCmd.AddCommand(listCmd)
}

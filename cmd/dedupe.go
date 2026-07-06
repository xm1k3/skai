package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/xm1k3/skai/internal/index"
)

var dedupeDryRun bool

var dedupeCmd = &cobra.Command{
	Use:   "dedupe",
	Short: "Remove duplicate skills from the index by content hash",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		idx, err := index.Load()
		if err != nil {
			return err
		}
		groups := map[string][]index.Skill{}
		for _, s := range idx.Skills {
			groups[s.ContentHash] = append(groups[s.ContentHash], s)
		}
		keep := map[string]bool{}
		removed := 0
		var hashes []string
		for hash, group := range groups {
			if len(group) > 1 {
				hashes = append(hashes, hash)
			}
		}
		sort.Strings(hashes)
		for _, hash := range hashes {
			group := groups[hash]
			sort.Slice(group, func(i, j int) bool {
				if group[i].Source != group[j].Source {
					return group[i].Source < group[j].Source
				}
				return group[i].Path < group[j].Path
			})
			fmt.Printf("%s\n", group[0].QualifiedName())
			for _, dup := range group[1:] {
				fmt.Printf("  duplicate: %s (%s)\n", dup.QualifiedName(), dup.Path)
				removed++
			}
			keep[group[0].Source+"|"+group[0].Path] = true
		}
		if removed == 0 {
			fmt.Println("No duplicates found")
			return nil
		}
		if dedupeDryRun {
			fmt.Printf("\n%d duplicates found, rerun without --dry-run to remove them from the index\n", removed)
			return nil
		}
		var deduped []index.Skill
		seen := map[string]bool{}
		for _, s := range idx.Skills {
			if len(groups[s.ContentHash]) > 1 {
				if !keep[s.Source+"|"+s.Path] {
					continue
				}
				if seen[s.ContentHash] {
					continue
				}
				seen[s.ContentHash] = true
			}
			deduped = append(deduped, s)
		}
		idx.Skills = deduped
		if err := index.Save(idx); err != nil {
			return err
		}
		fmt.Printf("\nRemoved %d duplicates, %d skills remain in the index\n", removed, len(deduped))
		return nil
	},
}

func init() {
	dedupeCmd.Flags().BoolVar(&dedupeDryRun, "dry-run", false, "report duplicates without modifying the index")
	rootCmd.AddCommand(dedupeCmd)
}

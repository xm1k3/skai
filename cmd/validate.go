package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xm1k3/skai/internal/index"
	"github.com/xm1k3/skai/internal/skill"
)

var validateCmd = &cobra.Command{
	Use:   "validate [skill]",
	Short: "Validate skill frontmatter, referenced paths and length",
	Long:  "Validate a single skill by name, a local skill directory, or the whole index when no argument is given.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		type entry struct {
			label string
			dir   string
		}
		var entries []entry
		if len(args) == 1 {
			if info, err := os.Stat(args[0]); err == nil && info.IsDir() {
				entries = append(entries, entry{args[0], args[0]})
			} else {
				idx, err := index.Load()
				if err != nil {
					return err
				}
				matches := idx.Find(args[0])
				if len(matches) == 0 {
					return fmt.Errorf("skill %q not found in index and is not a local directory", args[0])
				}
				for _, s := range matches {
					dir, err := s.LocalDir()
					if err != nil {
						return err
					}
					entries = append(entries, entry{s.QualifiedName(), dir})
				}
			}
		} else {
			idx, err := index.Load()
			if err != nil {
				return err
			}
			for _, s := range idx.Skills {
				dir, err := s.LocalDir()
				if err != nil {
					return err
				}
				entries = append(entries, entry{s.QualifiedName(), dir})
			}
		}
		errors := 0
		warnings := 0
		for _, e := range entries {
			issues := skill.Validate(e.dir)
			if len(issues) == 0 {
				if len(args) == 1 {
					fmt.Printf("%s: OK\n", e.label)
				}
				continue
			}
			fmt.Println(e.label)
			for _, issue := range issues {
				fmt.Printf("  %-5s %s\n", issue.Severity, issue.Message)
				if issue.Severity == skill.SeverityError {
					errors++
				} else {
					warnings++
				}
			}
		}
		fmt.Printf("\nValidated %d skills: %d errors, %d warnings\n", len(entries), errors, warnings)
		if errors > 0 {
			return fmt.Errorf("validation failed with %d errors", errors)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

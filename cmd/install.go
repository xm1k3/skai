package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xm1k3/skai/internal/index"
	"github.com/xm1k3/skai/internal/installer"
)

var (
	installTarget   string
	installProject  bool
	installPersonal bool
	installLink     bool
	installYes      bool
	installForce    bool
)

var installCmd = &cobra.Command{
	Use:   "install <skill>",
	Short: "Install a skill into a target environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !installer.ValidTarget(installTarget) {
			return fmt.Errorf("invalid or missing --target %q, valid targets are claude-code, codex, web", installTarget)
		}
		if installProject && installPersonal {
			return fmt.Errorf("--project and --personal are mutually exclusive")
		}
		scope := installer.ScopePersonal
		if installProject {
			scope = installer.ScopeProject
		}
		if installTarget == installer.TargetWeb {
			if installPersonal {
				return fmt.Errorf("target web only supports project exports")
			}
			if installLink {
				return fmt.Errorf("--link is not supported for target web")
			}
			scope = installer.ScopeProject
		}
		idx, err := index.Load()
		if err != nil {
			return err
		}
		matches := idx.Find(args[0])
		if len(matches) == 0 {
			return fmt.Errorf("skill %q not found in index", args[0])
		}
		if len(matches) > 1 {
			var names []string
			for _, m := range matches {
				names = append(names, m.QualifiedName())
			}
			return fmt.Errorf("skill name %q is ambiguous, use one of: %s", args[0], strings.Join(names, ", "))
		}
		s := matches[0]
		srcDir, err := s.LocalDir()
		if err != nil {
			return err
		}
		if _, err := os.Stat(srcDir); err != nil {
			return fmt.Errorf("source checkout missing at %s, run skai sync first", srcDir)
		}
		dest, err := installer.Destination(installTarget, scope, s.Name)
		if err != nil {
			return err
		}
		records, err := installer.LoadManifest()
		if err != nil {
			return err
		}
		if existing, ok := installer.FindRecord(records, s.Name, installTarget, scope); ok && existing.Source != s.Source && !installForce {
			return fmt.Errorf("skill %q is already installed from source %q, use --force to replace it", s.Name, existing.Source)
		}
		if _, err := os.Lstat(dest); err == nil {
			if _, ok := installer.FindRecord(records, s.Name, installTarget, scope); !ok && !installForce {
				return fmt.Errorf("%s already exists and is not managed by skai, use --force to overwrite", dest)
			}
		}
		if s.ClaudeCodeOnly && installTarget != installer.TargetClaudeCode {
			fmt.Println("Warning: this skill uses Claude Code specific frontmatter and may not work on the selected target")
		}
		fmt.Printf("Skill:                       %s (%s)\n", s.Name, s.Source)
		fmt.Printf("Destination:                 %s\n", dest)
		fmt.Printf("Risk level:                  %s\n", s.RiskLevel())
		fmt.Printf("Network calls:               %s\n", boolLabel(s.NetworkCalls))
		fmt.Printf("Destructive ops:             %s\n", boolLabel(s.DestructiveOps))
		fmt.Printf("Confirms before destructive: %s\n", boolLabel(s.ConfirmsBeforeDestructive))
		fmt.Printf("Has scripts:                 %s\n", boolLabel(s.HasScripts))
		if !installYes && !confirm("Proceed with install? [y/N]: ") {
			return fmt.Errorf("installation aborted")
		}
		if err := os.RemoveAll(dest); err != nil {
			return err
		}
		mode := "copied"
		switch {
		case installTarget == installer.TargetWeb:
			if err := installer.ZipDir(srcDir, s.Name, dest); err != nil {
				return err
			}
			mode = "exported"
		case installLink:
			if err := installer.Symlink(srcDir, dest); err != nil {
				return err
			}
			mode = "linked"
		default:
			if err := installer.CopyDir(srcDir, dest); err != nil {
				return err
			}
		}
		records = installer.ReplaceRecord(records, installer.Record{
			Name:        s.Name,
			Source:      s.Source,
			Target:      installTarget,
			Scope:       scope,
			Path:        dest,
			Link:        installLink,
			InstalledAt: time.Now().UTC(),
		})
		if err := installer.SaveManifest(records); err != nil {
			return err
		}
		fmt.Printf("Installed %s (%s) to %s\n", s.Name, mode, dest)
		return nil
	},
}

func init() {
	installCmd.Flags().StringVar(&installTarget, "target", "", "install target (claude-code, codex, web)")
	installCmd.Flags().BoolVar(&installProject, "project", false, "install into the current project")
	installCmd.Flags().BoolVar(&installPersonal, "personal", false, "install into the personal directory (default)")
	installCmd.Flags().BoolVar(&installLink, "link", false, "symlink instead of copying so future syncs propagate")
	installCmd.Flags().BoolVar(&installYes, "yes", false, "skip the confirmation prompt")
	installCmd.Flags().BoolVar(&installForce, "force", false, "replace an existing installation")
	installCmd.MarkFlagRequired("target")
	rootCmd.AddCommand(installCmd)
}

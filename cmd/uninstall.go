package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xm1k3/skai/internal/installer"
)

var (
	uninstallTarget   string
	uninstallProject  bool
	uninstallPersonal bool
	uninstallYes      bool
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <skill>",
	Short: "Remove an installed skill from a target environment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !installer.ValidTarget(uninstallTarget) {
			return fmt.Errorf("invalid or missing --target %q, valid targets are claude-code, codex, web", uninstallTarget)
		}
		if uninstallProject && uninstallPersonal {
			return fmt.Errorf("--project and --personal are mutually exclusive")
		}
		scope := installer.ScopePersonal
		if uninstallProject || uninstallTarget == installer.TargetWeb {
			scope = installer.ScopeProject
		}
		name := args[0]
		records, err := installer.LoadManifest()
		if err != nil {
			return err
		}
		dest := ""
		if rec, ok := installer.FindRecord(records, name, uninstallTarget, scope); ok {
			dest = rec.Path
		} else {
			dest, err = installer.Destination(uninstallTarget, scope, name)
			if err != nil {
				return err
			}
		}
		if _, err := os.Lstat(dest); err != nil {
			return fmt.Errorf("skill %q is not installed at %s", name, dest)
		}
		if !uninstallYes && !confirm(fmt.Sprintf("Remove %s? [y/N]: ", dest)) {
			return fmt.Errorf("uninstall aborted")
		}
		if err := os.RemoveAll(dest); err != nil {
			return err
		}
		records = installer.RemoveRecord(records, name, uninstallTarget, scope)
		if err := installer.SaveManifest(records); err != nil {
			return err
		}
		fmt.Printf("Removed %s\n", dest)
		return nil
	},
}

func init() {
	uninstallCmd.Flags().StringVar(&uninstallTarget, "target", "", "install target (claude-code, codex, web)")
	uninstallCmd.Flags().BoolVar(&uninstallProject, "project", false, "uninstall from the current project")
	uninstallCmd.Flags().BoolVar(&uninstallPersonal, "personal", false, "uninstall from the personal directory (default)")
	uninstallCmd.Flags().BoolVar(&uninstallYes, "yes", false, "skip the confirmation prompt")
	uninstallCmd.MarkFlagRequired("target")
	rootCmd.AddCommand(uninstallCmd)
}

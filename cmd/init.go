package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xm1k3/skai/internal/config"
)

var initForce bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create the skai config directory with default sources",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.SourcesPath()
		if err != nil {
			return err
		}
		if _, err := os.Stat(path); err == nil && !initForce {
			return fmt.Errorf("config already exists at %s, use --force to overwrite", path)
		}
		if err := config.Save(config.Default()); err != nil {
			return err
		}
		sourcesDir, err := config.SourcesDir()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
			return err
		}
		fmt.Printf("Created %s with %d default sources\n", path, len(config.Default().Sources))
		fmt.Println("Run skai sync to clone the sources and build the catalog")
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite an existing config")
	rootCmd.AddCommand(initCmd)
}

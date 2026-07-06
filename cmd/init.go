package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/xm1k3/skai/internal/config"
)

var (
	initForce   bool
	initOffline bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create the skai config directory with default sources",
	Long:  "Download the default sources config from the skai repository and write it to ~/.skai/sources.yaml. Use --offline to write the sources bundled with this binary instead.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.SourcesPath()
		if err != nil {
			return err
		}
		if _, err := os.Stat(path); err == nil && !initForce {
			return fmt.Errorf("config already exists at %s, use --force to overwrite", path)
		}

		var count int
		var origin string
		if initOffline {
			if err := config.Save(config.Default()); err != nil {
				return err
			}
			count = len(config.Default().Sources)
			origin = "bundled defaults"
		} else {
			data, cfg, err := config.FetchSources(config.DefaultSourcesURL)
			if err != nil {
				return fmt.Errorf("failed to download default sources: %w (use --offline to use the sources bundled with this binary)", err)
			}
			if err := config.WriteRaw(data); err != nil {
				return err
			}
			count = len(cfg.Sources)
			origin = config.DefaultSourcesURL
		}

		sourcesDir, err := config.SourcesDir()
		if err != nil {
			return err
		}
		if err := os.MkdirAll(sourcesDir, 0o755); err != nil {
			return err
		}
		fmt.Printf("Created %s with %d sources from %s\n", path, count, origin)
		fmt.Println("Run skai sync to clone the sources and build the catalog")
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite an existing config")
	initCmd.Flags().BoolVar(&initOffline, "offline", false, "use the sources bundled with this binary instead of downloading")
	rootCmd.AddCommand(initCmd)
}

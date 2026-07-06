package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/xm1k3/skai/internal/config"
	"github.com/xm1k3/skai/internal/gitutil"
	"github.com/xm1k3/skai/internal/index"
	"github.com/xm1k3/skai/internal/scanner"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Clone or update all enabled sources and rebuild the local index",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := gitutil.Available(); err != nil {
			return err
		}
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		sourcesDir, err := config.SourcesDir()
		if err != nil {
			return err
		}
		var all []index.Skill
		synced := 0
		for _, src := range cfg.Sources {
			if !src.Enabled {
				fmt.Printf("%-40s skipped (disabled)\n", src.Name)
				continue
			}
			dir := filepath.Join(sourcesDir, src.Name)
			action, err := gitutil.CloneOrPull(src.Repo, dir)
			if err != nil {
				fmt.Printf("%-40s failed: %v\n", src.Name, err)
				continue
			}
			skills := scanner.ScanSource(src, dir)
			fmt.Printf("%-40s %s, %d skills\n", src.Name, action, len(skills))
			all = append(all, skills...)
			synced++
		}
		idx := &index.Index{UpdatedAt: time.Now().UTC(), Skills: all}
		if err := index.Save(idx); err != nil {
			return err
		}
		indexPath, err := config.IndexPath()
		if err != nil {
			return err
		}
		fmt.Printf("\nIndexed %d skills from %d sources into %s\n", len(all), synced, indexPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

package config

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const DefaultSourcesURL = "https://raw.githubusercontent.com/xm1k3/skai/refs/heads/main/sources.yaml"

const maxSourcesSize = 1 << 20

type Source struct {
	Name    string `yaml:"name"`
	Repo    string `yaml:"repo"`
	Enabled bool   `yaml:"enabled"`
}

type Config struct {
	Sources []Source `yaml:"sources"`
}

func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".skai"), nil
}

func SourcesPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "sources.yaml"), nil
}

func SourcesDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "sources"), nil
}

func IndexPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "index.json"), nil
}

func ManifestPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "installed.json"), nil
}

func Default() Config {
	return Config{
		Sources: []Source{
			{Name: "composio-awesome-claude-skills", Repo: "https://github.com/ComposioHQ/awesome-claude-skills", Enabled: true},
			{Name: "sickn33-antigravity-awesome-skills", Repo: "https://github.com/sickn33/antigravity-awesome-skills", Enabled: true},
			{Name: "alirezarezvani-claude-skills", Repo: "https://github.com/alirezarezvani/claude-skills", Enabled: true},
			{Name: "behisecc-awesome-claude-skills", Repo: "https://github.com/BehiSecc/awesome-claude-skills", Enabled: true},
			{Name: "travisvn-awesome-claude-skills", Repo: "https://github.com/travisvn/awesome-claude-skills", Enabled: true},
		},
	}
}

func Load() (Config, error) {
	path, err := SourcesPath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, fmt.Errorf("sources config not found at %s, run skai init first", path)
		}
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("invalid sources config: %w", err)
	}
	if len(cfg.Sources) == 0 {
		return Config{}, fmt.Errorf("no sources configured in %s", path)
	}
	return cfg, nil
}

func Save(cfg Config) error {
	path, err := SourcesPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func FetchSources(url string) ([]byte, Config, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, Config{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, Config{}, fmt.Errorf("unexpected status %s fetching %s", resp.Status, url)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxSourcesSize))
	if err != nil {
		return nil, Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, Config{}, fmt.Errorf("invalid sources config downloaded from %s: %w", url, err)
	}
	if len(cfg.Sources) == 0 {
		return nil, Config{}, fmt.Errorf("downloaded config from %s has no sources", url)
	}
	return data, cfg, nil
}

func WriteRaw(data []byte) error {
	path, err := SourcesPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

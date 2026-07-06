package installer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/xm1k3/skai/internal/config"
)

type Record struct {
	Name        string    `json:"name"`
	Source      string    `json:"source"`
	Target      string    `json:"target"`
	Scope       string    `json:"scope"`
	Path        string    `json:"path"`
	Link        bool      `json:"link"`
	InstalledAt time.Time `json:"installed_at"`
}

func LoadManifest() ([]Record, error) {
	path, err := config.ManifestPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var records []Record
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func SaveManifest(records []Record) error {
	path, err := config.ManifestPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func FindRecord(records []Record, name, target, scope string) (Record, bool) {
	for _, r := range records {
		if r.Name == name && r.Target == target && r.Scope == scope {
			return r, true
		}
	}
	return Record{}, false
}

func ReplaceRecord(records []Record, rec Record) []Record {
	out := RemoveRecord(records, rec.Name, rec.Target, rec.Scope)
	return append(out, rec)
}

func RemoveRecord(records []Record, name, target, scope string) []Record {
	var out []Record
	for _, r := range records {
		if r.Name == name && r.Target == target && r.Scope == scope {
			continue
		}
		out = append(out, r)
	}
	return out
}

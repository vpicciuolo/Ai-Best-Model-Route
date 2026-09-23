package gateway

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/vpicciuolo/ai-best-model-route/internal/config"
)

// LoadConfig extracts this router's plugin config from a Bifrost config file,
// keeping the sidecar and plugin on one declarative source of truth.
func LoadConfig(path string) (config.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return config.Config{}, fmt.Errorf("read Bifrost config: %w", err)
	}
	var root struct {
		Plugins []struct {
			Name   string `json:"name"`
			Config any    `json:"config"`
		} `json:"plugins"`
	}
	if err := json.Unmarshal(data, &root); err != nil {
		return config.Config{}, fmt.Errorf("decode Bifrost config: %w", err)
	}
	for _, plugin := range root.Plugins {
		if plugin.Name == "ai-best-model-route" {
			return config.FromAny(plugin.Config)
		}
	}
	return config.Config{}, fmt.Errorf("ai-best-model-route plugin config is missing")
}

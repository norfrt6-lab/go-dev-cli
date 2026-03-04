package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Editor           string   `mapstructure:"editor"`
	DefaultHealthPath string  `mapstructure:"default_health_path"`
	LogRetentionDays int      `mapstructure:"log_retention_days"`
	MaxLogEntries    int      `mapstructure:"max_log_entries"`
	Theme            string   `mapstructure:"theme"`
	ScanPorts        []int    `mapstructure:"scan_ports"`
	DataDir          string   `mapstructure:"data_dir"`
}

func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	dataDir := filepath.Join(home, ".devx")

	return &Config{
		Editor:           "code",
		DefaultHealthPath: "/health",
		LogRetentionDays: 30,
		MaxLogEntries:    10000,
		Theme:            "dark",
		ScanPorts:        []int{3000, 3001, 4000, 5000, 5173, 5432, 6379, 8080, 8443, 9090},
		DataDir:          dataDir,
	}
}

func Load(cfgFile string) (*Config, error) {
	cfg := DefaultConfig()

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(cfg.DataDir)
	}

	viper.SetEnvPrefix("DEVX")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("editor", cfg.Editor)
	viper.SetDefault("default_health_path", cfg.DefaultHealthPath)
	viper.SetDefault("log_retention_days", cfg.LogRetentionDays)
	viper.SetDefault("max_log_entries", cfg.MaxLogEntries)
	viper.SetDefault("theme", cfg.Theme)
	viper.SetDefault("scan_ports", cfg.ScanPorts)
	viper.SetDefault("data_dir", cfg.DataDir)

	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		return nil, err
	}

	// Read config file (ignore if not found)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func GetEditor() string {
	editor := viper.GetString("editor")
	if editor == "" {
		if envEditor := os.Getenv("EDITOR"); envEditor != "" {
			return envEditor
		}
		return "code"
	}
	return editor
}

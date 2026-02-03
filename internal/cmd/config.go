package cmd

import (
	"fmt"
	"os"

	"github.com/dedene/raindrop-cli/internal/config"
)

type ConfigCmd struct {
	Path ConfigPathCmd `cmd:"" help:"Show configuration paths"`
	Get  ConfigGetCmd  `cmd:"" help:"Get configuration value"`
	Set  ConfigSetCmd  `cmd:"" help:"Set configuration value"`
}

type ConfigPathCmd struct{}

func (c *ConfigPathCmd) Run() error {
	dir, err := config.Dir()
	if err != nil {
		return fmt.Errorf("resolve config dir: %w", err)
	}

	configPath, err := config.ConfigPath()
	if err != nil {
		return fmt.Errorf("resolve config path: %w", err)
	}

	keyringDir, err := config.KeyringDir()
	if err != nil {
		return fmt.Errorf("resolve keyring dir: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Config dir:  %s\n", dir)
	fmt.Fprintf(os.Stdout, "Config file: %s\n", configPath)
	fmt.Fprintf(os.Stdout, "Keyring dir: %s\n", keyringDir)

	return nil
}

type ConfigGetCmd struct {
	Key string `arg:"" help:"Configuration key (default_output, timezone, oauth_port)"`
}

func (c *ConfigGetCmd) Run() error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	var value string

	switch c.Key {
	case "default_output":
		value = cfg.DefaultOutput
		if value == "" {
			value = "table"
		}
	case "timezone":
		value = cfg.Timezone
		if value == "" {
			value = "Local"
		}
	case "oauth_port":
		if cfg.OAuthPort == 0 {
			value = "8484"
		} else {
			value = fmt.Sprintf("%d", cfg.OAuthPort)
		}
	default:
		return fmt.Errorf("unknown config key: %s", c.Key)
	}

	fmt.Fprintln(os.Stdout, value)

	return nil
}

type ConfigSetCmd struct {
	Key   string `arg:"" help:"Configuration key"`
	Value string `arg:"" help:"Configuration value"`
}

func (c *ConfigSetCmd) Run() error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	switch c.Key {
	case "default_output":
		if c.Value != "table" && c.Value != "json" {
			return fmt.Errorf("invalid output mode: %s (must be 'table' or 'json')", c.Value)
		}

		cfg.DefaultOutput = c.Value
	case "timezone":
		cfg.Timezone = c.Value
	case "oauth_port":
		var port int
		if _, err := fmt.Sscanf(c.Value, "%d", &port); err != nil {
			return fmt.Errorf("invalid port: %s", c.Value)
		}

		cfg.OAuthPort = port
	default:
		return fmt.Errorf("unknown config key: %s", c.Key)
	}

	if err := config.WriteConfig(cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Set %s = %s\n", c.Key, c.Value)

	return nil
}

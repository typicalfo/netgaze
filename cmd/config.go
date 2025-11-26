package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config [action]",
	Short: "Manage netgaze configuration",
	Long: `Manage netgaze configuration including defaults and legacy API settings.

Actions:
  set-key    Legacy: store API key (not used in this version)
  show       Show current configuration
  clear      Clear all configuration`,
	Args: cobra.MaximumNArgs(1),
	RunE: runConfig,
}

func runConfig(cmd *cobra.Command, args []string) error {
	switch len(args) {
	case 0:
		return showConfig()
	case 1:
		switch args[0] {
		case "set-key":
			return setAPIKey()
		case "show":
			return showConfig()
		case "clear":
			return clearConfig()
		default:
			return fmt.Errorf("unknown config action: %s", args[0])
		}
	default:
		return cmd.Help()
	}
}

func setAPIKey() error {
	fmt.Print("Enter OpenRouter API key: ")
	var apiKey string
	_, err := fmt.Scanln(&apiKey)
	if err != nil {
		return fmt.Errorf("failed to read API key: %w", err)
	}

	if strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	config, err := loadConfig()
	if err != nil {
		return err
	}

	// API keys are not used in this version; accept and discard.
	_ = strings.TrimSpace(apiKey)
	return saveConfig(config)
}

func showConfig() error {
	config, err := loadConfig()
	if err != nil {
		return err
	}

	fmt.Println("netgaze configuration:")
	fmt.Printf("  Default Timeout: %s\n", config.DefaultTimeout)
	fmt.Printf("  Enable Port Scan: %v\n", config.EnablePorts)

	return nil
}

func clearConfig() error {
	configPath := getConfigPath()
	if configPath == "" {
		return fmt.Errorf("cannot determine config directory")
	}

	err := os.Remove(configPath)
	if os.IsNotExist(err) {
		fmt.Println("No configuration file found")
		return nil
	}
	if err != nil {
		return err
	}

	fmt.Println("Configuration cleared")
	return nil
}

func maskKey(key string) string {
	if key == "" {
		return "not set"
	}
	if len(key) <= 8 {
		return strings.Repeat("*", len(key))
	}
	return key[:4] + strings.Repeat("*", len(key)-8) + key[len(key)-4:]
}

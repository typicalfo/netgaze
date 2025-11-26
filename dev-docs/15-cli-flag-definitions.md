# Addendum 15 – CLI Flag Definitions (spf13/cobra)

**File**: `cmd/root.go`

**Root Command Structure**
```go
package cmd

import (
    "fmt"
    "os"
    "time"
    "github.com/spf13/cobra"
)

var (
    enablePorts bool
    output      string
    timeout     time.Duration
)

var rootCmd = &cobra.Command{
    Use:   "ng [flags] <ip|domain|url>",
    Short: "Fast network reconnaissance TUI",
    Long: `ng is a fast, single-binary TUI that runs common network
reconnaissance tools in parallel and presents results beautifully.

Mode:
  • Deterministic, offline templated output (no AI)

Examples:
  ng 1.1.1.1
  ng tui google.com --ports
  ng -ai suspicious.site --output md > report.md`,
    Args:         cobra.ExactArgs(1),
    SilenceUsage: true,
    RunE:         runNetgaze,
}

func init() {
    rootCmd.Flags().BoolVar(&enablePorts, "ports", false, 
        "Enable port scan of common ports (not enabled by default)")
    rootCmd.Flags().StringVar(&output, "output", "text", 
        "Output format: text, md, json, raw (for piping or automation)")
    rootCmd.Flags().StringVar(&output, "json", "", 
        "Legacy alias for --output json")
    rootCmd.Flags().DurationVar(&timeout, "timeout", 15*time.Second, 
        "Global timeout for all operations")
    
    // Hide the legacy --json flag from help but keep for compatibility
    rootCmd.Flags().MarkHidden("json")
    
    // Add subcommands
    rootCmd.AddCommand(configCmd)
    rootCmd.AddCommand(versionCmd)
}
```

**Config Command**
```go
var configCmd = &cobra.Command{
    Use:   "config [action]",
    Short: "Manage netgaze configuration",
    Long: `Manage netgaze configuration including API keys and settings.

Actions:
  set-key    Set OpenRouter API key for AI mode
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
```

**Version Command**
```go
var versionCmd = &cobra.Command{
    Use:   "version",
    Short: "Show version information",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Printf("netgaze v%s\n", version)
        fmt.Printf("Built: %s\n", buildDate)
        fmt.Printf("Go version: %s\n", runtime.Version())
    },
}
```

**Configuration Management**
```go
// Config file location: ~/.config/netgaze/config.json
type Config struct {
    OpenRouterAPIKey string `json:"openrouter_api_key"`
    DefaultTimeout   string `json:"default_timeout"`
    EnablePorts      bool   `json:"enable_ports"`
}

func getConfigPath() string {
    home, err := os.UserHomeDir()
    if err != nil {
        return ""
    }
    return filepath.Join(home, ".config", "netgaze", "config.json")
}

func loadConfig() (*Config, error) {
    configPath := getConfigPath()
    if configPath == "" {
        return &Config{}, nil
    }
    
    data, err := os.ReadFile(configPath)
    if os.IsNotExist(err) {
        return &Config{}, nil
    }
    if err != nil {
        return nil, err
    }
    
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    return &config, nil
}

func saveConfig(config *Config) error {
    configPath := getConfigPath()
    if configPath == "" {
        return fmt.Errorf("cannot determine config directory")
    }
    
    if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
        return err
    }
    
    data, err := json.MarshalIndent(config, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(configPath, data, 0644)
}
```

**API Key Setting**
```go
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
    
    config.OpenRouterAPIKey = strings.TrimSpace(apiKey)
    return saveConfig(config)
}
```

**Environment Variable Priority**
```go
func getAPIKey() (string, error) {
    // 1. Environment variable (highest priority)
    if key := os.Getenv("OPENROUTER_API_KEY"); key != "" {
        return key, nil
    }
    
    // 2. Config file
    config, err := loadConfig()
    if err != nil {
        return "", err
    }
    
    if config.OpenRouterAPIKey != "" {
        return config.OpenRouterAPIKey, nil
    }
    
    return "", fmt.Errorf("OPENROUTER_API_KEY required for AI mode (use 'netgaze config set-key' or set environment variable)")
}
```

**Flag Validation**
```go
func validateFlags() error {
    // Validate output format
    validOutputs := []string{"text", "md", "json", "raw"}
    valid := false
    for _, v := range validOutputs {
        if output == v {
            valid = true
            break
        }
    }
    if !valid {
        return fmt.Errorf("invalid output format: %s (valid: %s)", output, strings.Join(validOutputs, ", "))
    }
    
    // Validate timeout
    if timeout < 1*time.Second || timeout > 5*time.Minute {
        return fmt.Errorf("timeout must be between 1s and 5m")
    }
    
    // Check for incompatible flag combinations
    if noAgent && output == "json" && os.Stdout.IsTerminal() {
        return fmt.Errorf("JSON output in non-AI mode requires piping or file redirection")
    }
    
    return nil
}
```

**Help Examples**
```bash
# Basic usage (text output)
ng 8.8.8.8

# Interactive TUI mode
ng tui example.com --ports

# Markdown report
ng suspicious.site --output md > report.md

# JSON output for automation
ng 1.1.1.1 --output json

# Custom timeout
ng slow-server.com --timeout 30s

# Configure API key
ng config set-key

# Show configuration
ng config show
```

**Error Handling**
- Invalid arguments show help with specific error
- Missing required flags provide helpful suggestions
- Configuration errors are user-friendly
- All errors use cobra's error handling for consistent formatting
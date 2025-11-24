# Addendum 17 – Output Rendering Logic

**Responsibility**: Handle --output formats and piping detection for non-interactive use.

**File**: `internal/output/renderer.go`

**Output Format Detection**
```go
type Format string

const (
    FormatText Format = "text"
    FormatMD   Format = "markdown"
    FormatJSON Format = "json"
    FormatRaw  Format = "raw"
)

func detectFormat(outputFlag string, isPiped bool) Format {
    // Legacy --json flag support
    if outputFlag == "" {
        if isPiped {
            return FormatJSON
        }
        return FormatText
    }
    
    switch strings.ToLower(outputFlag) {
    case "json":
        return FormatJSON
    case "md", "markdown":
        return FormatMD
    case "raw":
        return FormatRaw
    default:
        return FormatText
    }
}
```

**Piping Detection**
```go
func isPiped() bool {
    stat, _ := os.Stdin.Stat()
    return (stat.Mode() & os.ModeCharDevice) == 0
}

func isOutputRedirected() bool {
    stat, _ := os.Stdout.Stat()
    return (stat.Mode() & os.ModeCharDevice) == 0
}
```

**Main Renderer**
```go
type Renderer struct {
    format    Format
    noAgent   bool
    isPiped   bool
    templates *templates.Templates
}

func New(formatFlag string, noAgent bool) *Renderer {
    piped := isOutputRedirected()
    format := detectFormat(formatFlag, piped)
    
    return &Renderer{
        format:    format,
        noAgent:   noAgent,
        isPiped:   piped,
        templates: templates.Load(),
    }
}

func (r *Renderer) Render(report *model.Report) error {
    switch r.format {
    case FormatJSON:
        return r.renderJSON(report)
    case FormatMD:
        return r.renderMarkdown(report)
    case FormatRaw:
        return r.renderRaw(report)
    default:
        return r.renderText(report)
    }
}
```

**JSON Output**
```go
func (r *Renderer) renderJSON(report *model.Report) error {
    data, err := json.MarshalIndent(report, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal JSON: %w", err)
    }
    
    fmt.Println(string(data))
    return nil
}
```

**Markdown Output**
```go
func (r *Renderer) renderMarkdown(report *model.Report) error {
    var buf strings.Builder
    
    if err := r.templates.SummaryMarkdown.Execute(&buf, report); err != nil {
        return fmt.Errorf("failed to execute markdown template: %w", err)
    }
    
    fmt.Print(buf.String())
    return nil
}
```

**Raw Output**
```go
func (r *Renderer) renderRaw(report *model.Report) error {
    var buf strings.Builder
    
    if err := r.templates.RawText.Execute(&buf, report); err != nil {
        return fmt.Errorf("failed to execute raw template: %w", err)
    }
    
    fmt.Print(buf.String())
    return nil
}
```

**Text Output (Terminal)**
```go
func (r *Renderer) renderText(report *model.Report) error {
    if r.isPiped || r.noAgent {
        // Piped or no-agent: use template without styling
        var buf strings.Builder
        if err := r.templates.SummaryText.Execute(&buf, report); err != nil {
            return fmt.Errorf("failed to execute text template: %w", err)
        }
        fmt.Print(buf.String())
        return nil
    }
    
    // Interactive: launch TUI
    return r.launchTUI(report)
}
```

**TUI Integration**
```go
func (r *Renderer) launchTUI(report *model.Report) error {
    if r.noAgent {
        // No-agent mode: show styled template output
        return r.renderStyledText(report)
    }
    
    // Full AI mode: launch bubbletea TUI
    model := ui.InitialModel(report.Target, false)
    model.SetReport(report)
    
    program := tea.NewProgram(model, tea.WithAltScreen())
    finalModel, err := program.Run()
    if err != nil {
        return fmt.Errorf("failed to start TUI: %w", err)
    }
    
    if finalModel.(ui.Model).ShouldExitWithError() {
        return fmt.Errorf("TUI exited with error")
    }
    
    return nil
}
```

**Styled Text Output**
```go
func (r *Renderer) renderStyledText(report *model.Report) error {
    var buf strings.Builder
    if err := r.templates.SummaryText.Execute(&buf, report); err != nil {
        return fmt.Errorf("failed to execute text template: %w", err)
    }
    
    // Apply lipgloss styling to template output
    styled := r.applyStyling(buf.String())
    fmt.Print(styled)
    return nil
}

func (r *Renderer) applyStyling(text string) string {
    lines := strings.Split(text, "\n")
    var styledLines []string
    
    for _, line := range lines {
        if strings.HasPrefix(line, report.Target) {
            // Header line
            styledLines = append(styledLines, 
                lipgloss.NewStyle().
                    Bold(true).
                    Foreground(lipgloss.Color("86")).
                    Render(line))
        } else if strings.Contains(line, "Warnings:") {
            // Warning line
            styledLines = append(styledLines,
                lipgloss.NewStyle().
                    Foreground(lipgloss.Color("208")).
                    Render(line))
        } else {
            styledLines = append(styledLines, line)
        }
    }
    
    return strings.Join(styledLines, "\n")
}
```

**Output Mode Decision Tree**
```go
func DetermineOutputMode(outputFlag string, noAgent bool) OutputMode {
    piped := isOutputRedirected()
    
    // Priority 1: Explicit --output flag
    if outputFlag != "" {
        if noAgent || piped {
            return OutputModeTemplate
        }
        return OutputModeTUI
    }
    
    // Priority 2: Piping detection
    if piped {
        return OutputModeJSON
    }
    
    // Priority 3: No-agent mode (first-class option)
    if noAgent {
        return OutputModeTemplate
    }
    
    // Default: Interactive TUI
    return OutputModeTUI
}

type OutputMode int

const (
    OutputModeTUI OutputMode = iota
    OutputModeTemplate
    OutputModeJSON
)
```

**Error Handling in Output**
```go
func (r *Renderer) RenderWithError(report *model.Report, err error) error {
    // Always show errors, regardless of format
    if report != nil {
        report.Errors["output"] = err.Error()
    }
    
    switch r.format {
    case FormatJSON:
        return r.renderJSONWithErrors(report, err)
    default:
        return r.renderTextWithErrors(report, err)
    }
}

func (r *Renderer) renderTextWithErrors(report *model.Report, err error) error {
    fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
    
    if report != nil {
        // Try to render partial report
        if renderErr := r.renderText(report); renderErr != nil {
            fmt.Fprintf(os.Stderr, "Failed to render report: %s\n", renderErr.Error())
        }
    }
    
    return err
}
```

**Performance Considerations**
- Template compilation at startup, not per-render
- Efficient string building with strings.Builder
- Minimal allocations in hot paths
- Early exit for JSON output (no template processing)
- Lazy TUI initialization only when needed

**Integration Points**
- Called from cmd/root.go after collection
- Receives complete Report struct
- Handles all format-specific logic
- Provides consistent error handling across formats
# Addendum 20 – TUI Template System & Styling Guide

**Responsibility**: Define comprehensive TUI template system using latest Bubble Tea and Lip Gloss patterns for consistent, professional UI.

**Files**: `internal/ui/styles.go`, `internal/ui/layout.go`, `internal/ui/components/`

## Core Styling System

**Color Palette (Adaptive)**
```go
package ui

import (
    "github.com/charmbracelet/lipgloss"
)

// Adaptive color palette for light/dark terminal support
var (
    // Primary colors
    Primary   = lipgloss.AdaptiveColor{Light: "#5B21B6", Dark: "#8B5CF6"} // Purple
    Secondary = lipgloss.AdaptiveColor{Light: "#1E40AF", Dark: "#3B82F6"} // Blue
    
    // Neutral colors
    Background = lipgloss.AdaptiveColor{Light: "#FAFAFA", Dark: "#0F0F0F"}
    Surface    = lipgloss.AdaptiveColor{Light: "#F5F5F5", Dark: "#1A1A1A"}
    Border     = lipgloss.AdaptiveColor{Light: "#E5E5E5", Dark: "#333333"}
    
    // Text colors
    Text       = lipgloss.AdaptiveColor{Light: "#1F2937", Dark: "#F9FAFB"}
    TextMuted  = lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"}
    TextBright = lipgloss.AdaptiveColor{Light: "#111827", Dark: "#FFFFFF"}
    
    // Status colors
    Success = lipgloss.AdaptiveColor{Light: "#059669", Dark: "#10B981"} // Green
    Warning = lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#F59E0B"} // Orange
    Error   = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#EF4444"} // Red
    
    // Accent colors
    Accent1 = lipgloss.AdaptiveColor{Light: "#7C3AED", Dark: "#A78BFA"} // Violet
    Accent2 = lipgloss.AdaptiveColor{Light: "#0891B2", Dark: "#06B6D4"} // Cyan
)
```

**Base Style Definitions**
```go
type Styles struct {
    // Layout styles
    App          lipgloss.Style
    Header       lipgloss.Style
    Content      lipgloss.Style
    Footer       lipgloss.Style
    
    // Navigation
    TabActive    lipgloss.Style
    TabInactive  lipgloss.Style
    TabContainer lipgloss.Style
    
    // Sections
    Section      lipgloss.Style
    SectionTitle lipgloss.Style
    Subsection   lipgloss.Style
    
    // Data display
    Table        lipgloss.Style
    TableHeader  lipgloss.Style
    TableRow     lipgloss.Style
    TableRowAlt  lipgloss.Style
    TableCell    lipgloss.Style
    
    // Status indicators
    StatusSuccess lipgloss.Style
    StatusWarning lipgloss.Style
    StatusError   lipgloss.Style
    
    // Interactive elements
    Input        lipgloss.Style
    InputFocused lipgloss.Style
    Button       lipgloss.Style
    ButtonActive lipgloss.Style
    
    // Text elements
    Title        lipgloss.Style
    Subtitle     lipgloss.Style
    Label        lipgloss.Style
    Value        lipgloss.Style
    Code         lipgloss.Style
    
    // Loading states
    Spinner      lipgloss.Style
    Progress     lipgloss.Style
}

func DefaultStyles() Styles {
    return Styles{
        // Main application container
        App: lipgloss.NewStyle().
            Padding(1, 2).
            Background(Background).
            Foreground(Text),
            
        // Header with app title and status
        Header: lipgloss.NewStyle().
            Bold(true).
            Foreground(TextBright).
            Padding(0, 1).
            MarginBottom(1),
            
        // Main content area
        Content: lipgloss.NewStyle().
            Padding(1, 0).
            MarginBottom(1),
            
        // Footer with help text
        Footer: lipgloss.NewStyle().
            Foreground(TextMuted).
            Padding(0, 1).
            Italic(true),
            
        // Tab navigation
        TabActive: lipgloss.NewStyle().
            Bold(true).
            Foreground(TextBright).
            Background(Primary).
            Padding(0, 2).
            MarginRight(1),
            
        TabInactive: lipgloss.NewStyle().
            Foreground(Text).
            Background(Surface).
            Border(lipgloss.NormalBorder(), false, true, false, true).
            BorderForeground(Border).
            Padding(0, 2).
            MarginRight(1),
            
        TabContainer: lipgloss.NewStyle().
            MarginBottom(1),
            
        // Content sections
        Section: lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(Border).
            Background(Surface).
            Padding(1, 2).
            MarginBottom(1),
            
        SectionTitle: lipgloss.NewStyle().
            Bold(true).
            Foreground(Primary).
            MarginBottom(1),
            
        Subsection: lipgloss.NewStyle().
            PaddingLeft(2).
            MarginTop(1),
            
        // Table styling
        Table: lipgloss.NewStyle().
            Border(lipgloss.NormalBorder()).
            BorderForeground(Border).
            MarginBottom(1),
            
        TableHeader: lipgloss.NewStyle().
            Bold(true).
            Foreground(Primary).
            Align(lipgloss.Center).
            Padding(0, 1),
            
        TableRow: lipgloss.NewStyle().
            Padding(0, 1),
            
        TableRowAlt: lipgloss.NewStyle().
            Background(Surface).
            Padding(0, 1),
            
        TableCell: lipgloss.NewStyle().
            Padding(0, 1),
            
        // Status indicators
        StatusSuccess: lipgloss.NewStyle().
            Foreground(Success).
            Bold(true),
            
        StatusWarning: lipgloss.NewStyle().
            Foreground(Warning).
            Bold(true),
            
        StatusError: lipgloss.NewStyle().
            Foreground(Error).
            Bold(true),
            
        // Input elements
        Input: lipgloss.NewStyle().
            Border(lipgloss.NormalBorder()).
            BorderForeground(Border).
            Background(Background).
            Foreground(Text).
            Padding(0, 1).
            Width(60),
            
        InputFocused: lipgloss.NewStyle().
            Border(lipgloss.NormalBorder()).
            BorderForeground(Primary).
            Background(Background).
            Foreground(Text).
            Padding(0, 1).
            Width(60),
            
        Button: lipgloss.NewStyle().
            Background(Surface).
            Foreground(Text).
            Border(lipgloss.NormalBorder()).
            BorderForeground(Border).
            Padding(0, 2),
            
        ButtonActive: lipgloss.NewStyle().
            Background(Primary).
            Foreground(TextBright).
            Bold(true).
            Padding(0, 2),
            
        // Text elements
        Title: lipgloss.NewStyle().
            Bold(true).
            Foreground(TextBright).
            MarginBottom(1),
            
        Subtitle: lipgloss.NewStyle().
            Foreground(TextMuted).
            MarginBottom(1),
            
        Label: lipgloss.NewStyle().
            Bold(true).
            Foreground(Text).
            MarginRight(1),
            
        Value: lipgloss.NewStyle().
            Foreground(Text),
            
        Code: lipgloss.NewStyle().
            Background(Surface).
            Foreground(Accent2).
            Padding(0, 1).
            SetFontStyle(lipgloss.SingletonStyle{}), // Monospace if available
            
        // Loading states
        Spinner: lipgloss.NewStyle().
            Foreground(Primary),
            
        Progress: lipgloss.NewStyle().
            Foreground(Primary).
            Bold(true),
    }
}
```

## Layout System

**Responsive Layout Manager**
```go
package ui

import (
    "strings"
    "github.com/charmbracelet/lipgloss"
)

type Layout struct {
    width  int
    height int
    styles Styles
}

func NewLayout(width, height int) *Layout {
    return &Layout{
        width:  width,
        height: height,
        styles: DefaultStyles(),
    }
}

func (l *Layout) RenderHeader(title, status string) string {
    titleStyle := l.styles.Header.Render(title)
    statusStyle := l.styles.Footer.Render(status)
    
    return lipgloss.JoinHorizontal(lipgloss.Right, titleStyle, statusStyle)
}

func (l *Layout) RenderTabs(tabs []string, active int) string {
    var tabViews []string
    
    for i, tab := range tabs {
        if i == active {
            tabViews = append(tabViews, l.styles.TabActive.Render(tab))
        } else {
            tabViews = append(tabViews, l.styles.TabInactive.Render(tab))
        }
    }
    
    return l.styles.TabContainer.Render(
        lipgloss.JoinHorizontal(lipgloss.Left, tabViews...),
    )
}

func (l *Layout) RenderSection(title, content string) string {
    titleRendered := l.styles.SectionTitle.Render(title)
    contentRendered := l.styles.Subsection.Render(content)
    
    return l.styles.Section.Render(
        lipgloss.JoinVertical(lipgloss.Left, titleRendered, contentRendered),
    )
}

func (l *Layout) RenderKeyValuePairs(pairs map[string]string) string {
    var rows []string
    
    for key, value := range pairs {
        label := l.styles.Label.Render(key + ":")
        val := l.styles.Value.Render(value)
        rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left, label, val))
    }
    
    return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (l *Layout) RenderTable(headers []string, rows [][]string) string {
    // Calculate column widths
    colWidths := make([]int, len(headers))
    for i, header := range headers {
        colWidths[i] = len(header)
        for _, row := range rows {
            if i < len(row) && len(row[i]) > colWidths[i] {
                colWidths[i] = len(row[i])
            }
        }
    }
    
    // Render header
    var headerCells []string
    for i, header := range headers {
        style := l.styles.TableHeader.Width(colWidths[i])
        headerCells = append(headerCells, style.Render(header))
    }
    headerRow := lipgloss.JoinHorizontal(lipgloss.Left, headerCells...)
    
    // Render data rows
    var dataRows []string
    for rowIndex, row := range rows {
        var cells []string
        for colIndex, cell := range row {
            var style lipgloss.Style
            if rowIndex%2 == 0 {
                style = l.styles.TableRow
            } else {
                style = l.styles.TableRowAlt
            }
            
            if colIndex < len(colWidths) {
                style = style.Width(colWidths[colIndex])
            }
            
            cells = append(cells, style.Render(cell))
        }
        dataRows = append(dataRows, lipgloss.JoinHorizontal(lipgloss.Left, cells...))
    }
    
    return l.styles.Table.Render(
        lipgloss.JoinVertical(lipgloss.Left, headerRow, dataRows...),
    )
}

func (l *Layout) RenderFooter(help []string) string {
    helpText := strings.Join(help, " | ")
    return l.styles.Footer.Render(helpText)
}

func (l *Layout) PlaceContent(content string) string {
    return lipgloss.Place(
        l.width, l.height,
        lipgloss.Center, lipgloss.Center,
        content,
        lipgloss.WithWhitespaceChars(" "),
        lipgloss.WithWhitespaceForeground(Surface),
    )
}
```

## Component Templates

**Data Display Components**
```go
// Network information display
func (l *Layout) NetworkInfo(report *model.Report) string {
    pairs := map[string]string{
        "IPv4": strings.Join(report.IPv4, ", "),
        "IPv6": strings.Join(report.IPv6, ", "),
        "PTR":  strings.Join(report.PTR, ", "),
    }
    
    // Remove empty entries
    for k, v := range pairs {
        if v == "" {
            delete(pairs, k)
        }
    }
    
    return l.RenderSection("Network Information", l.RenderKeyValuePairs(pairs))
}

// Geolocation display
func (l *Layout) Geolocation(report *model.Report) string {
    if report.Geo.Country == "" {
        return ""
    }
    
    pairs := map[string]string{
        "Location":  fmt.Sprintf("%s, %s, %s", report.Geo.City, report.Geo.Region, report.Geo.Country),
        "ISP":      report.Geo.ISP,
        "ASN":      report.Geo.ASN,
        "Org":      report.Geo.Org,
    }
    
    // Remove empty entries
    for k, v := range pairs {
        if v == "" {
            delete(pairs, k)
        }
    }
    
    return l.RenderSection("Geolocation", l.RenderKeyValuePairs(pairs))
}

// Connectivity status
func (l *Layout) Connectivity(report *model.Report) string {
    var status string
    var style lipgloss.Style
    
    if report.Ping.Success {
        status = fmt.Sprintf("Success (%s packets, %s loss)", 
            fmt.Sprintf("%d/%d", report.Ping.PacketsReceived, report.Ping.PacketsSent),
            fmt.Sprintf("%.1f%%", report.Ping.PacketLossPct))
        style = l.styles.StatusSuccess
    } else {
        status = "Failed: " + report.Ping.Error
        style = l.styles.StatusError
    }
    
    pairs := map[string]string{
        "Ping Status": style.Render(status),
    }
    
    if report.Ping.Success {
        pairs["RTT"] = fmt.Sprintf("Avg %s, Min %s, Max %s", 
            report.Ping.AvgRtt, report.Ping.MinRtt, report.Ping.MaxRtt)
    }
    
    return l.RenderSection("Connectivity", l.RenderKeyValuePairs(pairs))
}

// Services and ports
func (l *Layout) Services(report *model.Report) string {
    var pairs map[string]string
    
    if len(report.Ports.Open) > 0 {
        pairs = map[string]string{
            "Open Ports": strings.Join(
                func() []string {
                    var ports []string
                    for _, p := range report.Ports.Open {
                        ports = append(ports, fmt.Sprintf("%d", p))
                    }
                    return ports
                }(), ", "),
        }
    }
    
    if report.TLS.Subject != "" {
        if pairs == nil {
            pairs = make(map[string]string)
        }
        
        var status string
        var style lipgloss.Style
        
        if report.TLS.Expired {
            status = "Expired"
            style = l.styles.StatusWarning
        } else if report.TLS.SelfSigned {
            status = "Self-signed"
            style = l.styles.StatusWarning
        } else {
            status = "Valid"
            style = l.styles.StatusSuccess
        }
        
        pairs["TLS Certificate"] = fmt.Sprintf("%s (%s)", report.TLS.CommonName, style.Render(status))
        pairs["Expires"] = report.TLS.NotAfter
    }
    
    if len(pairs) == 0 {
        return ""
    }
    
    return l.RenderSection("Services", l.RenderKeyValuePairs(pairs))
}

// Error display
func (l *Layout) Errors(report *model.Report) string {
    if len(report.Errors) == 0 {
        return ""
    }
    
    var errorList []string
    for collector, error := range report.Errors {
        errorStyle := l.styles.StatusError.Render("Error")
        errorList = append(errorList, 
            fmt.Sprintf("%s: %s", collector, errorStyle.Render(error)))
    }
    
    return l.RenderSection("Warnings", strings.Join(errorList, "\n"))
}
```

## Integration with Bubble Tea Model

**Updated Model with Layout System**
```go
type Model struct {
    // Core data
    target    string
    report    *model.Report
    startTime time.Time
    
    // State management
    state     State
    currentTab Tab
    noAgent   bool
    
    // UI components
    spinner   spinner.Model
    viewport  viewport.Model
    textInput textinput.Model
    
    // Layout system
    layout    *Layout
    styles    Styles
    
    // Future interactive features (not implemented)
    messages  []ChatMessage
    streaming bool
    
    // Error handling
    fatalError error
}

func InitialModel(target string, noAgent bool) Model {
    // Initialize components
    s := spinner.New()
    s.Spinner = spinner.Points
    s.Style = DefaultStyles().Spinner
    
    ti := textinput.New()
    ti.Placeholder = "Ask about the network data..."
    ti.Focus()
    ti.CharLimit = 200
    ti.Width = 60
    
    vp := viewport.New(80, 20)
    
    return Model{
        target:    target,
        startTime: time.Now(),
        state:     StateCollecting,
        currentTab: TabSummary,
        noAgent:   noAgent,
        spinner:   s,
        textInput: ti,
        viewport:  vp,
        styles:    DefaultStyles(),
        messages:  []ChatMessage{},
    }
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Update layout on window resize
    if msg, ok := msg.(tea.WindowSizeMsg); ok {
        m.layout = NewLayout(msg.Width, msg.Height)
    }
    
    // Rest of update logic...
    return m, nil
}

func (m Model) View() string {
    if m.state == StateQuitting {
        return ""
    }
    
    // Initialize layout if not set
    if m.layout == nil {
        m.layout = NewLayout(80, 24)
    }
    
    // Build layout components
    header := m.layout.RenderHeader(
        fmt.Sprintf("netgaze: %s", m.target),
        m.getStatusText(),
    )
    
    tabs := m.layout.RenderTabs(
        []string{"Summary", "Raw Data", "Ask"},
        int(m.currentTab),
    )
    
    content := m.renderCurrentTab()
    
    footer := m.layout.RenderFooter(m.getHelpText())
    
    // Compose final layout
    return m.styles.App.Render(
        lipgloss.JoinVertical(
            lipgloss.Left,
            header,
            tabs,
            content,
            footer,
        ),
    )
}
```

## Usage Guidelines

**Consistency Requirements**
1. **Use adaptive colors** for all user-facing elements
2. **Apply consistent spacing** with margin/padding utilities
3. **Follow the component hierarchy** (App -> Section -> Subsection -> Content)
4. **Maintain responsive design** that adapts to terminal size
5. **Use semantic styling** (StatusSuccess for success states, etc.)

**Customization Points**
- **Color schemes**: Modify the adaptive color palette
- **Border styles**: Choose from lipgloss border options
- **Typography**: Adjust font weights and styles
- **Spacing**: Customize padding and margins
- **Layout**: Modify component arrangement

**Performance Considerations**
- **Cache rendered content** when possible
- **Minimize style allocations** in hot paths
- **Use efficient string building** for complex layouts
- **Profile memory usage** with large datasets

This template system ensures consistent, professional appearance across the entire netgaze application while maintaining flexibility for future enhancements.
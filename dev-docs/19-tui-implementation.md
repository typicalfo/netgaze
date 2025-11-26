# Addendum 19 – TUI Implementation with Code Samples

**Responsibility**: Complete TUI implementation using Charmbracelet ecosystem with production-ready code samples.

**Files**: `internal/ui/model.go`, `internal/ui/view.go`, `internal/ui/components/`

## Core TUI Architecture

**Model Structure**
```go
package ui

import (
    "time"
    "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/spinner"
    "github.com/charmbracelet/bubbles/viewport"
    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/table"
    "github.com/charmbracelet/lipgloss"
)

type State int

const (
    StateCollecting State = iota
    StateComplete
    StateAsking
    StateQuitting
)

type Tab int

const (
    TabSummary Tab = iota
    TabRawData
    TabAsk
)

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
    table     table.Model
    
    // AI mode specific
    // Placeholder for future integrations
    messages  []ChatMessage
    streaming bool
    messages  []ChatMessage
    streaming bool
    
    // Error handling
    fatalError error
    
    // Styling
    styles     Styles
}

type ChatMessage struct {
    Role    string // "user" or "assistant"
    Content string
    Time    time.Time
}

type Styles struct {
    Header           lipgloss.Style
    ActiveTab        lipgloss.Style
    InactiveTab      lipgloss.Style
    Section          lipgloss.Style
    SectionTitle     lipgloss.Style
    Error            lipgloss.Style
    Success          lipgloss.Style
    Warning          lipgloss.Style
    Footer           lipgloss.Style
}
```

**Initial Model Setup**
```go
func InitialModel(target string, noAgent bool) Model {
    // Initialize spinner
    s := spinner.New()
    s.Spinner = spinner.Points
    s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
    
    // Initialize text input for AI chat
    ti := textinput.New()
    ti.Placeholder = "Ask about the network data..."
    ti.Focus()
    ti.CharLimit = 200
    ti.Width = 60
    
    // Initialize viewport for scrollable content
    vp := viewport.New(80, 20)
    vp.Style = lipgloss.NewStyle().Padding(1)
    
    // Initialize table for raw data
    tbl := table.New(
        table.WithColumns([]table.Column{
            {Title: "Property", Width: 20},
            {Title: "Value", Width: 50},
        }),
        table.WithFocused(true),
        table.WithHeight(15),
    )
    
    return Model{
        target:    target,
        startTime: time.Now(),
        state:     StateCollecting,
        currentTab: TabSummary,
        noAgent:   noAgent,
        spinner:   s,
        textInput: ti,
        viewport:  vp,
        table:     tbl,
        messages:  []ChatMessage{},
        styles:    DefaultStyles(),
    }
}
```

**Styling System**
```go
func DefaultStyles() Styles {
    return Styles{
        Header: lipgloss.NewStyle().
            Bold(true).
            Foreground(lipgloss.Color("86")).
            Padding(0, 1),
            
        ActiveTab: lipgloss.NewStyle().
            Padding(0, 2).
            MarginRight(1).
            Background(lipgloss.Color("62")).
            Foreground(lipgloss.Color("230")).
            Bold(true),
            
        InactiveTab: lipgloss.NewStyle().
            Padding(0, 2).
            MarginRight(1).
            Background(lipgloss.Color("238")).
            Foreground(lipgloss.Color("245")),
            
        Section: lipgloss.NewStyle().
            Border(lipgloss.RoundedBorder()).
            BorderForeground(lipgloss.Color("86")).
            Padding(0, 1).
            MarginBottom(1),
            
        SectionTitle: lipgloss.NewStyle().
            Bold(true).
            Foreground(lipgloss.Color("86")).
            MarginBottom(1),
            
        Error: lipgloss.NewStyle().
            Foreground(lipgloss.Color("196")).
            Bold(true),
            
        Success: lipgloss.NewStyle().
            Foreground(lipgloss.Color("46")).
            Bold(true),
            
        Warning: lipgloss.NewStyle().
            Foreground(lipgloss.Color("208")).
            Bold(true),
            
        Footer: lipgloss.NewStyle().
            Foreground(lipgloss.Color("243")).
            Padding(0, 1),
    }
}
```

## Update Function with Message Handling

**Main Update Loop**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKeyMsg(msg)
        
    case collector.CompleteMsg:
        return m.handleCollectionComplete(msg)
        
    // StreamMsg handling removed – no AI integration in this version
        return m.handleAgentStream(msg)
        
    // ErrorMsg handling removed – no AI integration in this version
        return m.handleAgentError(msg)
        
    case tea.WindowSizeMsg:
        return m.handleWindowResize(msg)
        
    case spinner.TickMsg:
        if m.state == StateCollecting {
            m.spinner, cmd = m.spinner.Update(msg)
            return m, cmd
        }
    }
    
    // Update child components
    return m.updateChildComponents(msg)
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.Type {
    case tea.KeyCtrlC, tea.KeyQ:
        m.state = StateQuitting
        return m, tea.Quit
        
    case tea.Key1, tea.Key2, tea.Key3:
        if m.state == StateComplete {
            return m.switchTab(msg.Type)
        }
        
    case tea.KeyEnter:
        if m.state == StateComplete && m.currentTab == TabAsk && !m.noAgent {
            return m.handleUserQuestion()
        }
    }
    
    return m, nil
}

func (m Model) switchTab(keyType tea.KeyType) (tea.Model, tea.Cmd) {
    switch keyType {
    case tea.Key1:
        m.currentTab = TabSummary
    case tea.Key2:
        m.currentTab = TabRawData
    case tea.Key3:
        if !m.noAgent {
            m.currentTab = TabAsk
            m.textInput.Focus()
        }
    }
    return m, nil
}
```

## View Function with Layout

**Main View Rendering**
```go
func (m Model) View() string {
    if m.state == StateQuitting {
        return ""
    }
    
    // Build layout components
    header := m.renderHeader()
    tabs := m.renderTabs()
    content := m.renderContent()
    footer := m.renderFooter()
    
    // Compose final layout
    return lipgloss.JoinVertical(
        lipgloss.Left,
        header,
        tabs,
        content,
        footer,
    )
}

func (m Model) renderHeader() string {
    title := m.styles.Header.Render(fmt.Sprintf("netgaze: %s", m.target))
    
    elapsed := time.Since(m.startTime)
    elapsedStr := fmt.Sprintf("Elapsed: %.1fs", elapsed.Seconds())
    
    if m.state == StateCollecting {
        elapsedStr = fmt.Sprintf("%s %s", m.spinner.View(), elapsedStr)
    }
    
    elapsedStyled := lipgloss.NewStyle().
        Foreground(lipgloss.Color("243")).
        Render(elapsedStr)
    
    return lipgloss.JoinHorizontal(lipgloss.Right, title, elapsedStyled)
}

func (m Model) renderTabs() string {
    tabs := []string{"Summary", "Raw Data"}
    if !m.noAgent {
        tabs = append(tabs, "Ask")
    }
    
    var tabViews []string
    for i, tab := range tabs {
        if Tab(i) == m.currentTab {
            tabViews = append(tabViews, m.styles.ActiveTab.Render(tab))
        } else {
            tabViews = append(tabViews, m.styles.InactiveTab.Render(tab))
        }
    }
    
    return lipgloss.JoinHorizontal(lipgloss.Left, tabViews...)
}
```

## Tab Content Rendering

**Summary Tab**
```go
func (m Model) renderSummary() string {
    if m.noAgent {
        return m.renderTemplateSummary()
    }
    
    return m.renderAISummary()
}

func (m Model) renderTemplateSummary() string {
    if m.report == nil {
        return "Collecting data..."
    }
    
    var sections []string
    
    // Basic info section
    sections = append(sections, m.renderBasicInfo())
    
    // Network section
    sections = append(sections, m.renderNetworkSection())
    
    // Connectivity section
    sections = append(sections, m.renderConnectivitySection())
    
    // Services section
    sections = append(sections, m.renderServicesSection())
    
    // Warnings section
    if len(m.report.Errors) > 0 {
        sections = append(sections, m.renderWarnings())
    }
    
    return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderBasicInfo() string {
    if m.report == nil {
        return ""
    }
    
    title := m.styles.SectionTitle.Render("Basic Information")
    
    var content strings.Builder
    if len(m.report.IPv4) > 0 {
        content.WriteString(fmt.Sprintf("IPv4: %s\n", strings.Join(m.report.IPv4, ", ")))
    }
    if len(m.report.IPv6) > 0 {
        content.WriteString(fmt.Sprintf("IPv6: %s\n", strings.Join(m.report.IPv6, ", ")))
    }
    if m.report.Geo.Country != "" {
        content.WriteString(fmt.Sprintf("Location: %s, %s, %s\n", 
            m.report.Geo.City, m.report.Geo.Region, m.report.Geo.Country))
    }
    
    return m.styles.Section.Render(
        lipgloss.JoinVertical(lipgloss.Left, title, content.String()),
    )
}
```

**Raw Data Tab with Table**
```go
func (m Model) renderRawData() string {
    if m.report == nil {
        return "No data available"
    }
    
    // Update table with current data
    m.updateRawDataTable()
    
    return m.styles.Section.Render(
        lipgloss.JoinVertical(
            lipgloss.Left,
            m.styles.SectionTitle.Render("Raw Network Data"),
            m.table.View(),
        ),
    )
}

func (m *Model) updateRawDataTable() {
    var rows []table.Row
    
    // DNS information
    if len(m.report.IPv4) > 0 {
        rows = append(rows, table.Row{"IPv4 Addresses", strings.Join(m.report.IPv4, ", ")})
    }
    if len(m.report.IPv6) > 0 {
        rows = append(rows, table.Row{"IPv6 Addresses", strings.Join(m.report.IPv6, ", ")})
    }
    if len(m.report.PTR) > 0 {
        rows = append(rows, table.Row{"Reverse DNS", strings.Join(m.report.PTR, ", ")})
    }
    
    // Geolocation
    if m.report.Geo.Country != "" {
        rows = append(rows, table.Row{"Country", m.report.Geo.Country})
        rows = append(rows, table.Row{"City", m.report.Geo.City})
        rows = append(rows, table.Row{"ISP", m.report.Geo.ISP})
        rows = append(rows, table.Row{"ASN", m.report.Geo.ASN})
    }
    
    // Ping results
    if m.report.Ping.Success {
        rows = append(rows, table.Row{"Ping Success", "Yes"})
        rows = append(rows, table.Row{"Packet Loss", fmt.Sprintf("%.1f%%", m.report.Ping.PacketLossPct)})
        rows = append(rows, table.Row{"Average RTT", m.report.Ping.AvgRtt})
    } else {
        rows = append(rows, table.Row{"Ping Success", "No"})
        rows = append(rows, table.Row{"Ping Error", m.report.Ping.Error})
    }
    
    // Port scan results
    if len(m.report.Ports.Open) > 0 {
        rows = append(rows, table.Row{"Open Ports", strings.Join(
            func() []string {
                var ports []string
                for _, p := range m.report.Ports.Open {
                    ports = append(ports, fmt.Sprintf("%d", p))
                }
                return ports
            }(), ", ",
        )})
    }
    
    m.table.SetRows(rows)
}
```

**Ask Tab with Chat Interface**
```go
func (m Model) renderAsk() string {
    if m.noAgent {
        return m.styles.Section.Render(
            "Interactive Ask tab is not available in this version.",
        )
    }
    
    // Chat history
    var chatContent strings.Builder
    for _, msg := range m.messages {
        style := lipgloss.NewStyle().Padding(0, 1).MarginBottom(1)
        
        if msg.Role == "user" {
            style = style.Background(lipgloss.Color("240")).Foreground(lipgloss.Color("230"))
            chatContent.WriteString(style.Render(fmt.Sprintf("You: %s", msg.Content)))
        } else {
            style = style.Background(lipgloss.Color("236")).Foreground(lipgloss.Color("250"))
            chatContent.WriteString(style.Render(fmt.Sprintf("AI: %s", msg.Content)))
        }
    }
    
    // Input area
    inputStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("86")).
        Padding(0, 1).
        MarginTop(1)
    
    input := inputStyle.Render(m.textInput.View())
    
    return lipgloss.JoinVertical(lipgloss.Left, chatContent.String(), input)
}
```

## Component Integration

**Footer with Help Text**
```go
func (m Model) renderFooter() string {
    var help []string
    
    if m.state == StateComplete {
        help = append(help, "1-3: tabs")
        if m.currentTab == TabAsk && !m.noAgent {
            help = append(help, "enter: ask")
        }
    }
    
    help = append(help, "q/ctrl+c: quit")
    
    helpText := strings.Join(help, " | ")
    return m.styles.Footer.Render(helpText)
}
```

**Error Handling Display**
```go
func (m Model) renderError() string {
    if m.fatalError == nil {
        return ""
    }
    
    return m.styles.Error.Render(
        fmt.Sprintf("Error: %s", m.fatalError.Error()),
    )
}
```

## Program Initialization

**Main Program Entry**
```go
func RunTUI(target string, noAgent bool, report *model.Report) error {
    model := InitialModel(target, noAgent)
    if report != nil {
        model.report = report
        model.state = StateComplete
    }
    
    program := tea.NewProgram(
        model,
        tea.WithAltScreen(),
        tea.WithMouseCellMotion(),
        tea.WithFPS(60),
    )
    
    finalModel, err := program.Run()
    if err != nil {
        return fmt.Errorf("failed to start TUI: %w", err)
    }
    
    if finalModel.(Model).ShouldExitWithError() {
        return fmt.Errorf("TUI exited with error")
    }
    
    return nil
}
```

## Documentation Reference

**Use Context7 MCP Server for Additional Documentation**

For more advanced patterns and component usage, use the context7 MCP server:

```bash
# Get Bubble Tea documentation
context7-mcp_get-library-docs /charmbracelet/bubbletea

# Get Bubbles components documentation  
context7-mcp_get-library-docs /charmbracelet/bubbles

# Get Lip Gloss styling documentation
context7-mcp_get-library-docs /charmbracelet/lipgloss
```

**Key Areas to Explore:**
- **Advanced animations**: Harmonica integration with Bubble Tea
- **Complex layouts**: Advanced JoinHorizontal/Vertical patterns
- **Custom components**: Building reusable UI components
- **Accessibility**: Screen reader support and high contrast modes
- **Performance**: Optimization for large datasets
- **Testing**: Unit testing TUI components

## Template Examples for End Users

**Text Template Output:**
```
8.8.8.8 - Network Intelligence Report
IPs: 8.8.8.8
Location: Mountain View, California, US (Google LLC)
ASN: AS15169 - Google LLC

Ping: 5/5 packets, 12.4ms avg
Open Ports: 53, 443
TLS: *.google.com (expires: 2025-03-15)

Duration: 2,341ms
```

**Styled Terminal Output:**
- Uses Lip Gloss for colors and formatting
- Purple headers for sections
- Green for successful data, red for errors
- Consistent spacing and borders
- Adaptive colors for light/dark terminals

**Design Principles:**
1. **Clarity**: Information hierarchy with clear visual separation
2. **Scannability**: Important data prominently displayed
3. **Consistency**: Uniform styling across all sections
4. **Accessibility**: High contrast and readable fonts
5. **Responsiveness**: Adapts to different terminal sizes

This implementation provides a professional, user-friendly TUI that showcases network data effectively while maintaining the performance and reliability requirements outlined in the planning documents.
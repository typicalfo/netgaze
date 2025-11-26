# Addendum 16 – TUI Component Breakdown & State Machine

**File**: `internal/ui/model.go`, `internal/ui/view.go`

**TUI Architecture (bubbletea + lipgloss)**

**State Machine**
```go
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
```

**Model Structure**
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
    
    // UI state
    spinner   spinner.Model
    viewport  viewport.Model
    textInput textinput.Model
    
    // Future interactive Ask tab (not implemented)
    messages  []ChatMessage
    streaming bool
    
    // Error handling
    fatalError error
}

type ChatMessage struct {
    Role    string // "user" or "assistant"
    Content string
    Time    time.Time
}
```

**Initial Model**
```go
func InitialModel(target string, noAgent bool) Model {
    s := spinner.New()
    s.Spinner = spinner.Points
    s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
    
    ti := textinput.New()
    ti.Placeholder = "Ask about the network data..."
    ti.Focus()
    
    return Model{
        target:    target,
        startTime: time.Now(),
        state:     StateCollecting,
        currentTab: TabSummary,
        noAgent:   noAgent,
        spinner:   s,
        textInput: ti,
        messages:  []ChatMessage{},
    }
}
```

**Update Function**
```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.Type {
        case tea.KeyCtrlC, tea.KeyQ:
            m.state = StateQuitting
            return m, tea.Quit
            
        case tea.Key1, tea.Key2, tea.Key3:
            if m.state == StateComplete {
                switch msg.Type {
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
            }
            
        case tea.KeyEnter:
            if m.state == StateComplete && m.currentTab == TabAsk && !m.noAgent {
                if m.textInput.Value() != "" {
                    cmd = m.askQuestion(m.textInput.Value())
                    m.textInput.SetValue("")
                }
            }
        }
        
    case collector.CompleteMsg:
        m.report = msg.Report
        m.state = StateComplete
        if !m.noAgent {
            cmd = m.generateSummary()
        }
        
    // StreamMsg handling removed – no AI integration in this version
        m.messages = append(m.messages, ChatMessage{
            Role:    "assistant",
            Content: msg.Content,
            Time:    time.Now(),
        })
        m.streaming = msg.Continues
        if msg.Continues {
            cmd = m.waitForStream()
        }
        
    // ErrorMsg handling removed – no AI integration in this version
        m.fatalError = msg.Err
        m.state = StateComplete
    }
    
    // Update child components
    if m.state == StateCollecting {
        m.spinner, cmd = m.spinner.Update(msg)
    }
    if m.currentTab == TabAsk {
        m.textInput, cmd = m.textInput.Update(msg)
    }
    
    return m, cmd
}
```

**View Function**
```go
func (m Model) View() string {
    if m.state == StateQuitting {
        return ""
    }
    
    // Header
    header := m.renderHeader()
    
    // Content based on state
    var content string
    switch m.state {
    case StateCollecting:
        content = m.renderCollecting()
    case StateComplete:
        content = m.renderComplete()
    }
    
    // Footer
    footer := m.renderFooter()
    
    return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}
```

**Header Component**
```go
func (m Model) renderHeader() string {
    title := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("230")).
        Padding(0, 1).
        Render(fmt.Sprintf("netgaze: %s", m.target))
    
    elapsed := time.Since(m.startTime)
    elapsedStr := lipgloss.NewStyle().
        Foreground(lipgloss.Color("243")).
        Render(fmt.Sprintf("Elapsed: %.1fs", elapsed.Seconds()))
    
    if m.state == StateCollecting {
        elapsedStr = lipgloss.JoinHorizontal(lipgloss.Left, 
            m.spinner.View(), 
            elapsedStr,
        )
    }
    
    return lipgloss.JoinHorizontal(lipgloss.Right, title, elapsedStr)
}
```

**Tab Navigation**
```go
func (m Model) renderTabs() string {
    tabs := []string{"Summary", "Raw Data"}
    if !m.noAgent {
        tabs = append(tabs, "Ask")
    }
    
    var tabViews []string
    for i, tab := range tabs {
        style := lipgloss.NewStyle().
            Padding(0, 2).
            MarginRight(1)
        
        if Tab(i) == m.currentTab {
            style = style.
                Background(lipgloss.Color("62")).
                Foreground(lipgloss.Color("230"))
        } else {
            style = style.
                Background(lipgloss.Color("238")).
                Foreground(lipgloss.Color("245"))
        }
        
        tabViews = append(tabViews, style.Render(tab))
    }
    
    return lipgloss.JoinHorizontal(lipgloss.Left, tabViews...)
}
```

**Summary Tab**
```go
func (m Model) renderSummary() string {
    if m.noAgent {
        return m.renderTemplateSummary()
    }
    
    return m.renderAISummary()
}

func (m Model) renderAISummary() string {
    if len(m.messages) == 0 {
        return "Generating summary..."
    }
    
    var content strings.Builder
    for _, msg := range m.messages {
        if msg.Role == "assistant" {
            content.WriteString(msg.Content)
        }
    }
    
    return lipgloss.NewStyle().
        Padding(1).
        Width(80).
        Render(content.String())
}
```

**Raw Data Tab**
```go
func (m Model) renderRawData() string {
    sections := []string{
        m.renderDNSSection(),
        m.renderGeoSection(),
        m.renderPingSection(),
        m.renderPortsSection(),
        m.renderTLSSection(),
        m.renderWhoisSection(),
        m.renderErrorsSection(),
    }
    
    return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderDNSSection() string {
    title := lipgloss.NewStyle().
        Bold(true).
        Foreground(lipgloss.Color("86")).
        Render("DNS Information")
    
    var content strings.Builder
    if len(m.report.IPv4) > 0 {
        content.WriteString(fmt.Sprintf("IPv4: %s\n", strings.Join(m.report.IPv4, ", ")))
    }
    if len(m.report.IPv6) > 0 {
        content.WriteString(fmt.Sprintf("IPv6: %s\n", strings.Join(m.report.IPv6, ", ")))
    }
    if len(m.report.PTR) > 0 {
        content.WriteString(fmt.Sprintf("PTR: %s\n", strings.Join(m.report.PTR, ", ")))
    }
    
    return lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("86")).
        Padding(0, 1).
        MarginBottom(1).
        Render(lipgloss.JoinVertical(lipgloss.Left, title, content.String()))
}
```

**Ask Tab (reserved for future interactive help)**
```go
func (m Model) renderAsk() string {
    if m.noAgent {
        return "Interactive Ask tab is not available in this version"
    }
    
    input := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        Padding(0, 1).
        Render(m.textInput.View())
    
    var chat strings.Builder
    for _, msg := range m.messages {
        style := lipgloss.NewStyle().Padding(0, 1)
        if msg.Role == "user" {
            style = style.Background(lipgloss.Color("240"))
        } else {
            style = style.Background(lipgloss.Color("236"))
        }
        chat.WriteString(style.Render(msg.Content))
    }
    
    return lipgloss.JoinVertical(lipgloss.Left, chat, input)
}
```

**Footer Component**
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
    
    helpText := lipgloss.NewStyle().
        Foreground(lipgloss.Color("243")).
        Render(strings.Join(help, " | "))
    
    return lipgloss.PlaceHorizontal(80, lipgloss.Right, helpText)
}
```

**Error Handling**
```go
func (m Model) renderError() string {
    if m.fatalError == nil {
        return ""
    }
    
    errorStyle := lipgloss.NewStyle().
        Background(lipgloss.Color("196")).
        Foreground(lipgloss.Color("230")).
        Padding(1).
        Bold(true)
    
    return errorStyle.Render(fmt.Sprintf("Error: %s", m.fatalError.Error()))
}
```

**Performance Considerations**
- Lazy rendering: only render visible tab
- Text wrapping for long content
- Efficient string building
- Minimal allocations in update loop
- Smooth spinner animation during collection
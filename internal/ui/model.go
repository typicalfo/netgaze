package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/typicalfo/netgaze/internal/model"
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
	state      State
	currentTab Tab
	noAgent    bool

	// UI components
	spinner   spinner.Model
	viewport  viewport.Model
	textInput textinput.Model
	table     table.Model

	// Layout system
	layout *Layout
	styles Styles

	// Collection options
	enablePorts bool
	timeout     time.Duration

	// AI mode specific (placeholder for future)
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

func InitialModel(target string, noAgent bool, enablePorts bool, timeout time.Duration) Model {
	// Initialize spinner
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = DefaultStyles().Spinner

	// Initialize text input for AI chat
	ti := textinput.New()
	ti.Placeholder = "Ask about network data..."
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 60

	// Initialize viewport for scrollable content
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().Padding(1)

	// Initialize table for raw data
	tbl := table.New(
		table.WithColumns([]table.Column{
			{Title: "Property", Width: 25},
			{Title: "Value", Width: 50},
		}),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	return Model{
		target:      target,
		startTime:   time.Now(),
		state:       StateCollecting,
		currentTab:  TabSummary,
		noAgent:     noAgent,
		spinner:     s,
		textInput:   ti,
		viewport:    vp,
		table:       tbl,
		styles:      DefaultStyles(),
		enablePorts: enablePorts,
		timeout:     timeout,
		messages:    []ChatMessage{},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.layout = NewLayout(msg.Width, msg.Height)
		return m, nil

	case spinner.TickMsg:
		if m.state == StateCollecting {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	// Update child components
	if m.currentTab == TabAsk {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	m.table, cmd = m.table.Update(msg)

	return m, cmd
}

func (m Model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.state = StateQuitting
		return m, tea.Quit

	case "1", "2", "3":
		if m.state == StateComplete {
			return m.switchTab(msg.String())
		}

	case "enter":
		if m.state == StateComplete && m.currentTab == TabAsk && !m.noAgent {
			if m.textInput.Value() != "" {
				// Add user message
				m.messages = append(m.messages, ChatMessage{
					Role:    "user",
					Content: m.textInput.Value(),
					Time:    time.Now(),
				})
				m.textInput.SetValue("")
			}
		}
	}

	return m, nil
}

func (m Model) switchTab(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "1":
		m.currentTab = TabSummary
	case "2":
		m.currentTab = TabRawData
	case "3":
		if !m.noAgent {
			m.currentTab = TabAsk
			m.textInput.Focus()
		}
	}
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
		m.noAgent,
	)

	content := m.renderCurrentTab()

	footer := m.layout.RenderFooter(strings.Fields(m.getHelpText()))

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

func (m Model) getStatusText() string {
	elapsed := time.Since(m.startTime)
	if m.state == StateCollecting {
		return fmt.Sprintf("%s Collecting... (%.1fs)", m.spinner.View(), elapsed.Seconds())
	}
	return fmt.Sprintf("Completed (%.1fs)", elapsed.Seconds())
}

func (m Model) getHelpText() string {
	var help []string

	if m.state == StateComplete {
		help = append(help, "1-3: tabs")
		if m.currentTab == TabAsk && !m.noAgent {
			help = append(help, "enter: ask")
		}
	}

	help = append(help, "q/ctrl+c: quit")
	return fmt.Sprintf("%s", help)
}

func (m Model) renderCurrentTab() string {
	switch m.currentTab {
	case TabSummary:
		return m.renderSummary()
	case TabRawData:
		return m.renderRawData()
	case TabAsk:
		return m.renderAsk()
	default:
		return ""
	}
}

func (m Model) renderSummary() string {
	if m.report == nil {
		return m.layout.RenderSection("Status", "Collecting network data...")
	}

	var sections []string

	// Basic info section
	if networkInfo := m.layout.NetworkInfo(m.report); networkInfo != "" {
		sections = append(sections, networkInfo)
	}

	// Geolocation section
	if geoInfo := m.layout.Geolocation(m.report); geoInfo != "" {
		sections = append(sections, geoInfo)
	}

	// Connectivity section
	if connectivity := m.layout.Connectivity(m.report); connectivity != "" {
		sections = append(sections, connectivity)
	}

	// Services section
	if services := m.layout.Services(m.report); services != "" {
		sections = append(sections, services)
	}

	// Errors section
	if errors := m.layout.Errors(m.report); errors != "" {
		sections = append(sections, errors)
	}

	if len(sections) == 0 {
		return m.layout.RenderSection("Status", "No data collected")
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) renderRawData() string {
	if m.report == nil {
		return "No data available"
	}

	// Update table with current data
	m.updateRawDataTable()

	// Convert table rows to strings
	var rowStrings []string
	for _, row := range m.table.Rows() {
		rowStrings = append(rowStrings, fmt.Sprintf("%s\t%s", row[0], row[1]))
	}

	return m.layout.RenderTable(
		[]string{"Property", "Value"},
		rowStrings,
	)
}

func (m Model) renderAsk() string {
	if m.noAgent {
		return "Ask tab is disabled in this mode."
	}

	// Chat history
	var chatContent string
	for _, msg := range m.messages {
		style := m.styles.Section.Padding(0, 1).MarginBottom(1)

		if msg.Role == "user" {
			style = style.Background(lipgloss.Color("240")).Foreground(lipgloss.Color("230"))
			chatContent += style.Render(fmt.Sprintf("You: %s", msg.Content))
		} else {
			style = style.Background(lipgloss.Color("236")).Foreground(lipgloss.Color("250"))
			chatContent += style.Render(fmt.Sprintf("AI: %s", msg.Content))
		}
	}

	// Input area
	input := m.styles.Input.Render(m.textInput.View())

	return lipgloss.JoinVertical(lipgloss.Left, chatContent, input)
}

func (m *Model) updateRawDataTable() {
	var rows []table.Row

	// Target information
	rows = append(rows, table.Row{"Target", m.report.Target})

	// DNS information
	if len(m.report.IPv4) > 0 {
		rows = append(rows, table.Row{"IPv4 Addresses", fmt.Sprintf("%v", m.report.IPv4)})
	}
	if len(m.report.IPv6) > 0 {
		rows = append(rows, table.Row{"IPv6 Addresses", fmt.Sprintf("%v", m.report.IPv6)})
	}
	if len(m.report.PTR) > 0 {
		rows = append(rows, table.Row{"Reverse DNS", fmt.Sprintf("%v", m.report.PTR)})
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
		if m.report.Ping.Error != "" {
			rows = append(rows, table.Row{"Ping Error", m.report.Ping.Error})
		}
	}

	// Port scan results
	if len(m.report.Ports.Open) > 0 {
		var ports []string
		for _, p := range m.report.Ports.Open {
			ports = append(ports, fmt.Sprintf("%d", p))
		}
		rows = append(rows, table.Row{"Open Ports", fmt.Sprintf("%v", ports)})
	}

	// TLS information
	if m.report.TLS.Subject != "" {
		rows = append(rows, table.Row{"TLS Subject", m.report.TLS.Subject})
		rows = append(rows, table.Row{"TLS Issuer", m.report.TLS.Issuer})
		rows = append(rows, table.Row{"TLS Common Name", m.report.TLS.CommonName})
		if len(m.report.TLS.AltNames) > 0 {
			rows = append(rows, table.Row{"TLS Alt Names", fmt.Sprintf("%v", m.report.TLS.AltNames)})
		}
		rows = append(rows, table.Row{"TLS Expires", m.report.TLS.NotAfter})
		if m.report.TLS.Expired {
			rows = append(rows, table.Row{"TLS Status", "Expired"})
		} else if m.report.TLS.SelfSigned {
			rows = append(rows, table.Row{"TLS Status", "Self-signed"})
		} else {
			rows = append(rows, table.Row{"TLS Status", "Valid"})
		}
	}

	// Duration
	if m.report.DurationMs > 0 {
		rows = append(rows, table.Row{"Duration (ms)", fmt.Sprintf("%d", m.report.DurationMs)})
	}

	m.table.SetRows(rows)
}

// SetReport updates the model with completed collection data
func (m *Model) SetReport(report *model.Report) {
	m.report = report
	m.state = StateComplete
}

// ShouldExitWithError returns true if the TUI should exit with an error
func (m Model) ShouldExitWithError() bool {
	return m.fatalError != nil
}

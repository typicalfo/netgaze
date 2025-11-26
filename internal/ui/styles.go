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

type Styles struct {
	// Layout styles
	App     lipgloss.Style
	Header  lipgloss.Style
	Content lipgloss.Style
	Footer  lipgloss.Style

	// Navigation
	TabActive    lipgloss.Style
	TabInactive  lipgloss.Style
	TabContainer lipgloss.Style

	// Sections
	Section      lipgloss.Style
	SectionTitle lipgloss.Style
	Subsection   lipgloss.Style

	// Data display
	Table       lipgloss.Style
	TableHeader lipgloss.Style
	TableRow    lipgloss.Style
	TableRowAlt lipgloss.Style
	TableCell   lipgloss.Style

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
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Label    lipgloss.Style
	Value    lipgloss.Style
	Code     lipgloss.Style

	// Loading states
	Spinner  lipgloss.Style
	Progress lipgloss.Style
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
			Align(lipgloss.Left).Bold(true).
			Padding(0, 2),

		TableRow: lipgloss.NewStyle().
			Padding(0, 2),

		TableRowAlt: lipgloss.NewStyle().
			Background(Surface).
			Padding(0, 2),

		TableCell: lipgloss.NewStyle().
			Padding(0, 2),

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
			MaxWidth(80).Align(lipgloss.Left),

		InputFocused: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(Primary).
			Background(Background).
			Foreground(Text).
			Padding(0, 1).
			MaxWidth(80).Align(lipgloss.Left),

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
			Padding(0, 2),

		// Loading states
		Spinner: lipgloss.NewStyle().
			Foreground(Primary),

		Progress: lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true),
	}
}

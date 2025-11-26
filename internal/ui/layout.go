package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/typicalfo/netgaze/internal/model"
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

func (l *Layout) RenderTabs(tabs []string, active int, noAgent bool) string {
	var tabViews []string

	// Filter tabs based on noAgent flag
	if noAgent && len(tabs) > 2 {
		tabs = tabs[:2] // Only show Summary and Raw Data
	}

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

func (l *Layout) RenderTable(headers []string, rows []string) string {
	if len(rows) == 0 {
		return l.styles.Section.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				l.styles.SectionTitle.Render("Data"),
				"No data available",
			),
		)
	}

	// Simple table rendering using lipgloss
	var tableRows []string

	// Header row
	var headerCells []string
	for _, header := range headers {
		headerCells = append(headerCells, l.styles.TableHeader.Render(header))
	}
	tableRows = append(tableRows, lipgloss.JoinHorizontal(lipgloss.Left, headerCells...))

	// Data rows
	for _, row := range rows {
		tableRows = append(tableRows, l.styles.TableRow.Render(row))
	}

	return l.styles.Table.Render(
		lipgloss.JoinVertical(lipgloss.Left, tableRows...),
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

	if len(pairs) == 0 {
		return ""
	}

	return l.RenderSection("Network Information", l.RenderKeyValuePairs(pairs))
}

// Geolocation display
func (l *Layout) Geolocation(report *model.Report) string {
	if report.Geo.Country == "" {
		return ""
	}

	pairs := map[string]string{
		"Location": fmt.Sprintf("%s, %s, %s", report.Geo.City, report.Geo.Region, report.Geo.Country),
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
		errorLabel := l.styles.StatusError.Render("Error")
		errorList = append(errorList,
			fmt.Sprintf("%s: %s %s", collector, errorLabel, error))
	}

	return l.RenderSection("Warnings", strings.Join(errorList, "\n"))
}

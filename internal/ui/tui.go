package ui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/typicalfo/netgaze/internal/model"
	"time"
)

// RunTUI starts the terminal user interface
func RunTUI(target string, noAgent bool, enablePorts bool, timeout time.Duration, report *model.Report) error {
	// Create initial model
	m := InitialModel(target, noAgent, enablePorts, timeout)

	// If report is already available, set it
	if report != nil {
		m.SetReport(report)
	}

	// Create and run program
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithFPS(60),
	)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}

	// Check if we should exit with an error
	if finalModel.(Model).ShouldExitWithError() {
		return fmt.Errorf("TUI exited with error")
	}

	return nil
}

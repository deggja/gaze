package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// lipgloss styles
var (
	appStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Width(50). // Set your desired width
			Render

	keywordStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Width(40). // Set your desired width
			Render

	selectedRowStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#5D3FD3")).
				Render

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("#333333")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Render

	tableStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Render
)

type podItem struct {
	Name      string
	Namespace string
	Keyword   string
}

type model struct {
	pods        []podItem
	cursor      int
	selectedPod podItem
	progressBar progress.Model
	loading     bool
}

type tickMsg time.Time

func tickEveryHalfSecond() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	// Start the progress bar immediately
	m.progressBar.SetPercent(0) // Reset progress to 0
	return tickEveryHalfSecond()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit // Handle quitting
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.pods)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selectedPod = m.pods[m.cursor]
		}

	case tickMsg:
		if m.loading {
			cmd = m.progressBar.IncrPercent(0.10) // Increment the progress bar
			return m, tea.Batch(cmd, tickEveryHalfSecond())
		}

	case progress.FrameMsg:
		// Update the progress bar model based on the frame message.
		var cmd tea.Cmd
		progressModel, cmd := m.progressBar.Update(msg)
		m.progressBar = progressModel.(progress.Model)

		if m.progressBar.Percent() >= 1.0 {
			m.loading = false
		}
		return m, cmd
		// Include other cases (e.g., window size adjustment) as necessary
	}

	// Return the model and any commands
	return m, cmd
}

func renderRow(pod podItem, isSelected bool, appWidth int) string {
	// Apply the app and keyword styles to the pod name and keyword
	app := appStyle(pod.Name)
	keyword := keywordStyle(pod.Keyword)

	// If the row is selected, apply the selectedRowStyle
	if isSelected {
		app = selectedRowStyle(app)
		keyword = selectedRowStyle(keyword)
	}

	return fmt.Sprintf("%s %s\n", app, keyword)
}

func (m model) View() string {
	if m.loading {
		// If still loading, show the progress bar.
		return fmt.Sprintf("\nCollecting log data..\n%s\n\nPress 'q' to quit.\n", m.progressBar.View())
	}

	// Once loading is complete, we render the table
	var b strings.Builder
	b.WriteString("Select the pod to inspect (press 'q' to quit):\n\n")

	appWidth := 0
	for _, pod := range m.pods {
		if len(pod.Name) > appWidth {
			appWidth = len(pod.Name)
		}
	}
	appWidth += 2 // Add some padding

	// Add table headers
	b.WriteString(headerStyle("Application") + " " + headerStyle("Keyword") + "\n")

	// Iterate over each pod and render the row
	for i, pod := range m.pods {
		// Use the renderRow function to format each row
		b.WriteString(renderRow(pod, m.cursor == i, appWidth))
	}

	// Return the rendered table
	return b.String()
}

func initialModel() model {
	pb := progress.New(progress.WithScaledGradient("#FDFF8C", "#FF7CCB"))
	pb.Width = 40
	pb.SetPercent(0) // Start with the progress at 0%

	return model{
		progressBar: pb,
		loading:     true,
	}
}

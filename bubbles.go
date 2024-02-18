package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"k8s.io/client-go/kubernetes"
)

// lipgloss styles
var (
	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Background(lipgloss.Color("#5D3FD3"))

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF"))
)

type podItem struct {
	Name       string
	Namespace  string
	Keyword    string
	LogContext string
}

type model struct {
	pods               []podItem
	cursor             int
	selectedPod        podItem
	progressBar        progress.Model
	loading            bool
	state              viewState
	selectedLogContext string
	clientset          *kubernetes.Clientset
}

type viewState int

const (
	viewListing viewState = iota
	viewLogDetails
)

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
			if m.state == viewListing {
				m.state = viewLogDetails
				m.selectedPod = m.pods[m.cursor]
				m.selectedLogContext = getLogDetailsForPod(m.clientset, m.selectedPod)
			}
		case "r", "R":
			if m.state == viewLogDetails {
				m.state = viewListing
				m.selectedLogContext = ""
			}
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

func renderRow(pod podItem, isSelected bool, appWidth, keywordWidth int) string {
	// Create the formatted application name and keyword
	appFormatted := fmt.Sprintf("%-*s", appWidth, pod.Name)
	keywordFormatted := fmt.Sprintf("%-*s", keywordWidth, pod.Keyword)

	// Combine the application name and keyword into one string
	row := appFormatted + " " + keywordFormatted

	if isSelected {
		// If the row is selected, apply the selectedRowStyle to the entire row
		row = selectedRowStyle.Render(row)
	}

	return row + "\n"
}

func (m model) View() string {
	if m.loading {
		// If still loading, show the progress bar.
		return fmt.Sprintf("\nCollecting log data..\n%s\n\nPress 'q' to quit.\n", m.progressBar.View())
	}

	var b strings.Builder

	switch m.state {
	case viewListing:
		b.WriteString("Select the pod to inspect (press 'q' to quit):\n\n")

		// Calculate the width for the application column
		appWidth := 0
		for _, pod := range m.pods {
			if len(pod.Name) > appWidth {
				appWidth = len(pod.Name)
			}
		}
		appWidth += 2 // Add some padding

		// Calculate the width for the keyword column
		keywordWidth := 0
		for _, pod := range m.pods {
			if len(pod.Keyword) > keywordWidth {
				keywordWidth = len(pod.Keyword)
			}
		}
		keywordWidth += 2 // Add some padding

		// Render the headers
		header := fmt.Sprintf("%-*s %-*s", appWidth, "Application", keywordWidth, "Keyword")
		b.WriteString(headerStyle.Render(header) + "\n")

		// Iterate over each pod and render the rows
		for i, pod := range m.pods {
			b.WriteString(renderRow(pod, m.cursor == i, appWidth, keywordWidth))
		}

	case viewLogDetails:
		// Render the detailed log view for the selected pod
		b.WriteString(fmt.Sprintf("Log details for %s:\n\n%s\n\nPress 'r' to return to the list.\n",
			m.selectedPod.Name, m.selectedLogContext))
	}

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

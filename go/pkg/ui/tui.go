package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ghchinoy/pandemonium/pkg/types"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AFBEE1")).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#3C3C3C"))

	agentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	worldStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EE6FF8")).
			Italic(true)
)

type model struct {
	events       chan types.Event
	log          []string
	agents       map[string]string
	remoteAgents map[string]bool
	viewport     viewport.Model
	ready        bool
}

type eventMsg types.Event

// NewTUI initializes a new Bubble Tea program for simulation observation.
func NewTUI(events chan types.Event) *tea.Program {
	m := &model{
		events:       events,
		agents:       make(map[string]string),
		remoteAgents: make(map[string]bool),
	}
	return tea.NewProgram(m, tea.WithAltScreen())
}

func (m *model) Init() tea.Cmd {
	return waitForEvent(m.events)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width-30, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 30
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

	case eventMsg:
		m.processEvent(types.Event(msg))
		m.viewport.SetContent(strings.Join(m.log, "\n"))
		m.viewport.GotoBottom()
		return m, waitForEvent(m.events)
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	sidebar := m.sidebarView()
	mainView := fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())

	return lipgloss.JoinHorizontal(lipgloss.Top, mainView, sidebar)
}

func waitForEvent(events chan types.Event) tea.Cmd {
	return func() tea.Msg {
		return eventMsg(<-events)
	}
}

func (m *model) processEvent(e types.Event) {
	var line string
	width := m.viewport.Width - 4

	switch e.Type {
	case types.EventSimStart:
		line = worldStyle.Render(fmt.Sprintf("✨ %s (start)", e.Message))
	case types.EventStepStart:
		line = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#555555")).Render(fmt.Sprintf("\n--- %s ---", e.Message))
	case types.EventObservation:
		prefix := agentStyle.Render(fmt.Sprintf("[%s]", e.Actor))
		text := lipgloss.NewStyle().Width(width).Render(fmt.Sprintf("%s (observation) %s", prefix, e.Message))
		line = text
		m.agents[e.Actor] = "Observing..."
		if data, ok := e.Data.(map[string]any); ok {
			if remote, ok := data["isRemote"].(bool); ok {
				m.remoteAgents[e.Actor] = remote
			}
		}
	case types.EventAction:
		prefix := agentStyle.Foreground(lipgloss.Color("#FFD700")).Render(fmt.Sprintf("[%s]", e.Actor))
		text := lipgloss.NewStyle().Width(width).Render(fmt.Sprintf("%s (action) %s", prefix, e.Message))
		line = text
		m.agents[e.Actor] = "Acting"
	case types.EventResolution:
		prefix := worldStyle.Bold(true).Render("[World]")
		text := lipgloss.NewStyle().Width(width).Render(fmt.Sprintf("%s (resolution) %s", prefix, e.Message))
		line = text
		// Clear statuses after a resolution
		for name := range m.agents {
			m.agents[name] = "Ready"
		}
	case types.EventThought:
		prefix := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Italic(true).Render(fmt.Sprintf("[%s]", e.Actor))
		text := lipgloss.NewStyle().Width(width).Render(fmt.Sprintf("%s (thought) %s", prefix, e.Message))
		line = text
		m.agents[e.Actor] = "Thinking..."
	case types.EventSimComplete:
		line = titleStyle.Render("Simulation Complete (end)")
	}
	if line != "" {
		m.log = append(m.log, line)
	}
}

func (m *model) headerView() string {
	title := titleStyle.Render("PANDEMONIUM SOCIAL SIMULATION")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m *model) footerView() string {
	info := infoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (m *model) sidebarView() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Width(26).Align(lipgloss.Center).Render("AGENTS"))
	sb.WriteString("\n\n")

	// Fixed order if we want consistency, or sorted
	var names []string
	for name := range m.agents {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		status := m.agents[name]
		displayName := name
		if m.remoteAgents[name] {
			displayName = name + " ☁️"
		}
		sb.WriteString(agentStyle.Render(fmt.Sprintf(" ● %s", displayName)))
		sb.WriteString("\n")

		// Wrap status text for sidebar
		wrappedStatus := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#777777")).
			Italic(true).
			Width(24).
			PaddingLeft(4).
			Render(status)

		sb.WriteString(wrappedStatus)
		sb.WriteString("\n\n")
	}
	return infoStyle.Width(28).Height(m.viewport.Height + 2).Render(sb.String())
}

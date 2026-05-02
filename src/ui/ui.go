package ui

import (
	"harry/session/src/session"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type Model struct {
	Sessions        []session.Session
	ViewSessions    []session.Session
	HideInactive    bool
	Search          string
	Cursor          int
	SelectedSession *session.Session
	Width           int
	Height          int
}

func InitialModel(sessions []session.Session) Model {
	return Model{
		Sessions:        sessions,
		ViewSessions:    session.FuzzySearch(sessions, ""),
		HideInactive:    false,
		Search:          "",
		Cursor:          0,
		SelectedSession: nil,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "ctrl+j", "down":
			m.Cursor++
		case "ctrl+k", "up":
			m.Cursor--
		case "ctrl+w":
			blankIndex := strings.Index(m.Search, " ")
			slashIndex := strings.Index(m.Search, "/")
			separator := " "
			if slashIndex > blankIndex {
				separator = "/"
			}
			parts := strings.Split(m.Search, separator)
			if len(parts) == 1 {
				m.Search = ""
			} else {
				m.Search = strings.Join(parts[:len(parts)-1], separator)
			}
		case "enter":
			m.SelectedSession = &m.ViewSessions[m.Cursor]
			return m, tea.Quit
		case "backspace":
			if len(m.Search) > 0 {
				m.Search = m.Search[:len(m.Search)-1]
			}
		case "ctrl+h":
			m.HideInactive = !m.HideInactive
		default:
			if msg.Key().Mod == 0 {
				m.Search += msg.Key().Text
			}
		}

		m.ViewSessions = filterSessions(m.Sessions, m.HideInactive)
		m.ViewSessions = session.FuzzySearch(m.ViewSessions, m.Search)

		if m.Cursor >= len(m.ViewSessions) {
			m.Cursor = len(m.ViewSessions) - 1
		} else if m.Cursor < 0 {
			m.Cursor = 0
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

func (m Model) View() tea.View {
	if m.Width == 0 || m.Height == 0 {
		return tea.NewView("")
	}

	var s strings.Builder
	bufferSpace := 4

	longestName := 0
	longestBranch := 0
	longestPath := 0
	for _, session := range m.Sessions {
		if len(session.Name) > longestName {
			longestName = len(session.Name)
		}
		if len(session.Branch) > longestBranch {
			longestBranch = len(session.Branch)
		}
		if len(session.Path) > longestPath {
			longestPath = len(session.Path)
		}
	}

	appWidth := 4 + longestName + bufferSpace + longestPath + bufferSpace + longestBranch

	xStart := max((m.Width-appWidth)/2, 0)
	yStart := max(m.Height/8, 1)

	s.WriteString(strings.Repeat("\n", yStart))

	searchBoxWidth := max(m.Width/3, 0)
	innerSearchBoxWidth := min(len(m.Search), searchBoxWidth-2)
	s.WriteString(strings.Repeat(" ", max(m.Width/3, 1)-1))
	s.WriteString("╭")
	s.WriteString(strings.Repeat("─", searchBoxWidth-2))
	s.WriteString("╮")
	s.WriteString("\n")
	s.WriteString(strings.Repeat(" ", max(m.Width/3, 1)-1))
	s.WriteString("│")
	s.WriteString(m.Search[0:innerSearchBoxWidth])
	if innerSearchBoxWidth < searchBoxWidth-2 {
		s.WriteString("█")
	}
	s.WriteString(strings.Repeat(" ", max(searchBoxWidth-innerSearchBoxWidth-3, 0)))
	s.WriteString("│")
	s.WriteString("\n")
	s.WriteString(strings.Repeat(" ", max(m.Width/3, 1)-1))
	s.WriteString("╰")
	s.WriteString(strings.Repeat("─", searchBoxWidth-2))
	s.WriteString("╯")
	s.WriteString("\n\n\n\n")

	name := "Name"
	path := "Path"
	branch := "Branch"
	s.WriteString(strings.Repeat(" ", xStart+4))
	s.WriteString(name)
	s.WriteString(strings.Repeat(" ", longestName+bufferSpace-len(name)))
	s.WriteString(path)
	s.WriteString(strings.Repeat(" ", longestPath+bufferSpace-len(path)))
	s.WriteString(branch)
	s.WriteString("\n\n")

	for i, session := range m.ViewSessions {
		s.WriteString(strings.Repeat(" ", xStart))
		if i == m.Cursor {
			s.WriteString("▸ ")
		} else {
			s.WriteString("  ")
		}
		if session.IsActive {
			s.WriteString("* ")
		} else {
			s.WriteString("  ")
		}
		nameBuffer := strings.Repeat(" ", longestName+bufferSpace-len(session.Name))
		pathBuffer := strings.Repeat(" ", longestPath+bufferSpace-len(session.Path))
		s.WriteString(session.Name)
		s.WriteString(nameBuffer)
		s.WriteString(session.Path)
		s.WriteString(pathBuffer)
		s.WriteString(session.Branch)
		s.WriteString("\n")
	}
	view := tea.NewView(s.String())
	view.AltScreen = true
	return view
}

func filterSessions(sessions []session.Session, hideInactive bool) []session.Session {
	if !hideInactive {
		return sessions
	}
	var filteredSessions []session.Session
	for _, session := range sessions {
		if session.IsActive {
			filteredSessions = append(filteredSessions, session)
		}
	}
	return filteredSessions
}

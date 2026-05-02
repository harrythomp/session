package ui

import (
	"fmt"
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
		case "ctrl+s":
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
	var s strings.Builder
	s.WriteString(m.Search)
	s.WriteString("\n\n")
	longestName := 0
	longestBranch := 0
	for _, session := range m.Sessions {
		if len(session.Name) > longestName {
			longestName = len(session.Name)
		}
		if len(session.Branch) > longestBranch {
			longestBranch = len(session.Branch)
		}
	}
	longestName += 4
	longestBranch += 4
	for i, session := range m.ViewSessions {
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
		nameBuffer := strings.Repeat(" ", longestName-len(session.Name))
		branchBuffer := strings.Repeat(" ", longestBranch-len(session.Branch))
		fmt.Fprintf(&s, "%s%s%s%s%s\n", session.Name, nameBuffer, session.Branch, branchBuffer, session.Path)
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

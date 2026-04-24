package session

import (
	"errors"
	"os/exec"
	"strings"
)

func findSessionsFromTmux() ([]Session, error) {
	names, err := listTmuxSessionsF("#{session_name}")
	if err != nil {
		return nil, err
	}
	paths, err := listTmuxSessionsF("#{session_path}")
	if err != nil {
		return nil, err
	}

	if len(names) != len(paths) {
		return nil, errors.New("Unable to read tmux sessions")
	}

	sessions := make([]Session, 0)

	for i, name := range names {
		sessions = append(sessions, Session{
			Name:     name,
			Path:     paths[i],
			IsActive: true,
		})
	}

	return sessions, nil
}

func listTmuxSessionsF(format string) ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", format)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	str := string(out)
	trimmed := strings.TrimSpace(str)
	lines := strings.Split(trimmed, "\n")
	return lines, nil
}

package session

import (
	"errors"
	"harry/session/src/config"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func findSessionsFromTmux() ([]Session, error) {
	names := listTmuxSessionsF("#{session_name}")
	paths := listTmuxSessionsF("#{session_path}")

	if len(names) != len(paths) {
		return nil, errors.New("Unable to read tmux sessions")
	}

	sessions := make([]Session, 0)

	for i, name := range names {
		branch, _ := getBranchFromRealPath(paths[i])
		sessions = append(sessions, Session{
			Name:        name,
			Path:        paths[i],
			ProjectPath: paths[i],
			Branch:      branch,
			IsActive:    true,
		})
	}

	return sessions, nil
}

func listTmuxSessionsF(format string) []string {
	cmd := exec.Command("tmux", "list-sessions", "-F", format)
	out, err := cmd.Output()
	if err != nil {
		return make([]string, 0)
	}
	str := string(out)
	trimmed := strings.TrimSpace(str)
	lines := strings.Split(trimmed, "\n")
	return lines
}

func AttachToSession(conf config.Config, session Session) error {
	if !session.IsActive {
		err := startNewTmuxSession(session)
		if err != nil {
			return err
		}
		err = sessionInit(conf, session)
		if err != nil {
			return err
		}
	}
	err := attachTmuxToSession(session)
	if err != nil {
		return err
	}
	return nil
}

func attachTmuxToSession(session Session) error {
	inSession := os.Getenv("TMUX") != ""
	if inSession {
		cmd := exec.Command("tmux", "switch", "-t", session.Name)
		_, err := cmd.Output()
		if err != nil {
			return err
		}
		return nil
	} else {
		cmd := exec.Command("tmux", "attach", "-t", session.Name)
		cmd.Stdin = os.Stdin
		_, err := cmd.Output()
		if err != nil {
			return err
		}
		return nil
	}
}

func startNewTmuxSession(session Session) error {
	cmd := exec.Command("tmux", "new-session", "-c", session.Path, "-s", session.Name, "-d")
	_, err := cmd.Output()
	return err
}

func sessionInit(conf config.Config, session Session) error {
	localScript := filepath.Join(session.Path, ".session")
	if found, err := tryInitScript(localScript, session); found {
		return err
	}

	globalScript := filepath.Join(conf.Location, "scripts", session.Name)
	if found, err := tryInitScript(globalScript, session); found {
		return err
	}

	projectLocalScript := filepath.Join(session.ProjectPath, ".session")
	if found, err := tryInitScript(projectLocalScript, session); found {
		return err
	}

	projectGlobalScript := filepath.Join(conf.Location, "scripts", filepath.Base(session.ProjectPath))
	if found, err := tryInitScript(projectGlobalScript, session); found {
		return err
	}

	defaultScript := filepath.Join(conf.Location, "default-session")
	if found, err := tryInitScript(defaultScript, session); found {
		return err
	}

	return nil
}

func tryInitScript(script string, session Session) (bool, error) {
	if file, err := os.Stat(script); err == nil && !file.IsDir() {
		_, err := exec.Command("tmux", "send-keys", "-t", session.Name+":1", script+" "+session.Name, "c-M").Output()
		if err != nil {
			return true, err
		}
		return true, nil
	}
	return false, nil
}

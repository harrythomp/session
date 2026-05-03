package session

import (
	"errors"
	"fmt"
	"harry/session/src/config"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var tmuxSessionDataMismatchError = errors.New("unable to read tmux sessions (length of session data doesn't match)")

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
		return nil, tmuxSessionDataMismatchError
	}

	sessions := make([]Session, 0, len(names))

	for i, name := range names {
		branch, _ := getBranchFromRealPath(paths[i]) // ignore error, branch is optional
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

func listTmuxSessionsF(format string) ([]string, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", format)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error when getting tmux sessions: %w", err)
	}
	str := string(out)
	trimmed := strings.TrimSpace(str)
	lines := strings.Split(trimmed, "\n")
	return lines, nil
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
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error when starting new tmux session: %w", err)
	}
	return nil
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
		err := exec.Command("tmux", "send-keys", "-t", session.Name+":1", script+" "+session.Name, "c-M").Run()
		if err != nil {
			return true, fmt.Errorf("error when running init script (%s): %w", script, err)
		}
		return true, nil
	}
	return false, nil
}

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

type TmuxSessionFinder struct{}

func (f TmuxSessionFinder) FindSessions() ([]Session, error) {
	tmuxSessions, err := findTmuxSessions()
	if err != nil {
		return nil, err
	}
	sessions := make([]Session, 0, len(tmuxSessions))
	for _, tmuxSession := range tmuxSessions {
		session := NewSessionFromWorkingPath(tmuxSession.Path, true)
		session.Name = tmuxSession.Name
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (f TmuxSessionFinder) MergeSessions(currentSessions []Session, newSessions []Session) []Session {
	sessionMap := make(map[string]Session, len(currentSessions))
	for _, session := range currentSessions {
		sessionMap[session.Name+":"+session.WorkingPath] = session
	}
	for _, session := range newSessions {
		sessionMap[session.Name+":"+session.WorkingPath] = session
	}
	mergedSessions := make([]Session, 0, len(sessionMap))
	for _, session := range sessionMap {
		mergedSessions = append(mergedSessions, session)
	}
	return mergedSessions
}

var tmuxSessionDataMismatchError = errors.New("unable to read tmux sessions (length of session data doesn't match)")

type TmuxSession struct {
	Name string
	Path string
}

func findTmuxSessions() ([]TmuxSession, error) {
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

	sessions := make([]TmuxSession, 0, len(names))

	for i, name := range names {
		sessions = append(sessions, TmuxSession{
			Name: name,
			Path: paths[i],
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

func AttachTmuxToSession(conf config.Config, session Session) error {
	if !session.IsActive {
		err := startNewTmuxSession(session)
		if err != nil {
			return err
		}
		err = tmuxSessionInit(conf, session)
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
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("error when switching to session %s: %v", session.Name, err)
		}
	} else {
		cmd := exec.Command("tmux", "attach", "-t", session.Name)
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("error when attaching to session %s: %v", session.Name, err)
		}
	}
	return nil
}

func startNewTmuxSession(session Session) error {
	cmd := exec.Command("tmux", "new-session", "-c", session.WorkingPath, "-s", session.Name, "-d")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error when starting new tmux session: %w", err)
	}
	return nil
}

func tmuxSessionInit(conf config.Config, session Session) error {
	localScript := filepath.Join(session.WorkingPath, ".session")
	if found, err := tryTmuxSessionInitScript(localScript, session); found {
		return err
	}

	globalScript := filepath.Join(conf.Location, "scripts", session.Name)
	if found, err := tryTmuxSessionInitScript(globalScript, session); found {
		return err
	}

	repositoryLocalScript := filepath.Join(session.RepositoryPath, ".session")
	if found, err := tryTmuxSessionInitScript(repositoryLocalScript, session); found {
		return err
	}

	repositoryGlobalScript := filepath.Join(conf.Location, "scripts", filepath.Base(session.RepositoryPath))
	if found, err := tryTmuxSessionInitScript(repositoryGlobalScript, session); found {
		return err
	}

	defaultScript := filepath.Join(conf.Location, "default-session")
	if found, err := tryTmuxSessionInitScript(defaultScript, session); found {
		return err
	}

	return nil
}

func tryTmuxSessionInitScript(script string, session Session) (bool, error) {
	if file, err := os.Stat(script); err == nil && !file.IsDir() {
		err := exec.Command("tmux", "send-keys", "-t", session.Name+":1", script+" "+session.Name, "c-M").Run()
		if err != nil {
			return true, fmt.Errorf("error when running init script (%s): %w", script, err)
		}
		return true, nil
	}
	return false, nil
}

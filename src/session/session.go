package session

import (
	"errors"
	"harry/session/src/config"
	"io/fs"
	"os"
	"slices"
	"strings"
)

type Session struct {
	Name     string
	Path     string
	IsActive bool
}

type sessionKey struct {
	name string
	path string
}

func (session Session) key() sessionKey {
	return sessionKey{
		name: session.Name,
		path: session.Path,
	}
}

func FindSessions(conf config.Config) ([]Session, error) {
	inactiveSessions := make([]Session, 0)
	for _, path := range conf.SearchPaths {
		sessions, err := findSessionsFromPath(path)
		if err != nil {
			return nil, err
		}
		inactiveSessions = append(inactiveSessions, sessions...)
	}

	for _, path := range conf.IncludePaths {
		dir, err := os.Stat(path)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if !dir.IsDir() {
			continue
		}
		inactiveSessions = append(inactiveSessions, Session{
			Name:     dir.Name(),
			Path:     path,
			IsActive: false,
		})
	}

	activeSessions, err := findSessionsFromTmux()
	if err != nil {
		return nil, err
	}

	sessions := mergeSessions(inactiveSessions, activeSessions)
	sortSessions(sessions)

	return sessions, nil
}

func sortSessions(sessions []Session) {
	slices.SortFunc(sessions, func(a, b Session) int {
		if a.IsActive && !b.IsActive {
			return -1
		} else if !a.IsActive && b.IsActive {
			return 1
		}
		return strings.Compare(a.Path+a.Name, b.Path+b.Name)
	})
}

func mergeSessions(sessionSlices ...[]Session) []Session {
	sessionMap := make(map[sessionKey]Session)

	for _, sessionSlice := range sessionSlices {
		for _, session := range sessionSlice {
			sessionMap[session.key()] = session
		}
	}

	mergedSessions := make([]Session, 0)
	for _, session := range sessionMap {
		mergedSessions = append(mergedSessions, session)
	}
	return mergedSessions
}

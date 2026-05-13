package session

import (
	"path/filepath"
	"sort"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

type Session struct {
	Name           string
	WorkingPath    string
	RepositoryPath string
	Branch         string
	IsActive       bool
}

type SessionKey struct{ Name, WorkingPath string }

func (s Session) UniqueKey() SessionKey {
	return SessionKey{Name: s.Name, WorkingPath: s.WorkingPath}
}

func (s *Session) SetName(name string) {
	s.Name = cleanTmuxName(name)
}

func NewSessionFromWorkingPath(path string) Session {
	session := Session{
		WorkingPath:    path,
		RepositoryPath: path,
		Branch:         "",
	}
	session.SetName(filepath.Base(path))

	worktrees, err := findGitWorktreesFromPath(path)

	// There are no worktrees, or there is only one worktree in the same directory
	if err != nil ||
		len(worktrees) == 0 ||
		(len(worktrees) == 1 &&
			(worktrees[0].Path == path)) {
		if len(worktrees) == 1 {
			session.Branch = worktrees[0].Branch
		}
		return session
	} else {
		for _, worktree := range worktrees {
			if worktree.Path == path {
				session.Branch = worktree.Branch
			} else if worktree.IsBare {
				session.RepositoryPath = worktree.Path
			}
		}
		return session
	}
}

type SessionFinder interface {
	FindSessions() ([]Session, error)
}

type MergeFunc func(currentSessions []Session, newSessions []Session) []Session

func MergeSessionsPreferActive(currentSessions []Session, newSessions []Session) []Session {
	sessionMap := make(map[SessionKey]Session, len(currentSessions))
	for _, session := range currentSessions {
		sessionMap[session.UniqueKey()] = session
	}
	for _, session := range newSessions {
		_, ok := sessionMap[session.UniqueKey()]
		if !ok || session.IsActive {
			sessionMap[session.UniqueKey()] = session
		}
	}
	mergedSessions := make([]Session, 0, len(sessionMap))
	for _, session := range sessionMap {
		mergedSessions = append(mergedSessions, session)
	}
	return mergedSessions
}

func FindSessions(sources []SessionFinder, mergeFunc MergeFunc) ([]Session, error) {
	var sessions []Session
	for _, source := range sources {
		sourceSessions, err := source.FindSessions()
		if err != nil {
			return nil, err
		}
		sessions = mergeFunc(sessions, sourceSessions)
	}
	return sessions, nil
}

func FuzzySearch(sessions []Session, search string) []Session {
	fuzzyStrings := make([]string, 0, len(sessions))
	for _, session := range sessions {
		fuzzyStrings = append(fuzzyStrings, session.WorkingPath+session.Name)
	}

	matches := fuzzy.RankFind(search, fuzzyStrings)
	sort.Sort(matches)

	fuzzySessions := make([]Session, 0, len(matches))
	for _, match := range matches {
		fuzzySessions = append(fuzzySessions, sessions[match.OriginalIndex])
	}

	return fuzzySessions
}

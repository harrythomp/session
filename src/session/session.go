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

func newSessionsFromRepositoryPath(path string, isActive bool) []Session {
	allWorktrees, err := findGitWorktreesFromPath(path)
	worktrees := make([]GitWorktree, 0, len(allWorktrees))
	for _, worktree := range allWorktrees {
		if worktree.IsBare {
			continue
		}
		worktrees = append(worktrees, worktree)
	}

	sessions := make([]Session, 0, len(worktrees))

	// There are no worktrees, or there is only one worktree in the same or a parent directory
	if err != nil ||
		len(worktrees) == 0 ||
		(len(worktrees) == 1 &&
			(worktrees[0].Path == path || len(worktrees[0].Path) < len(path))) {
		branch := ""
		if len(worktrees) == 1 {
			branch = worktrees[0].Branch
		}
		sessions = append(sessions, Session{
			Name:           filepath.Base(path),
			WorkingPath:    path,
			RepositoryPath: path,
			Branch:         branch,
			IsActive:       isActive,
		})
	} else {
		for _, worktree := range worktrees {
			sessions = append(sessions, Session{
				Name:           filepath.Base(path) + "[" + filepath.Base(worktree.Path) + "]",
				WorkingPath:    worktree.Path,
				RepositoryPath: path,
				Branch:         worktree.Branch,
				IsActive:       isActive,
			})
		}
	}
	return sessions
}

func NewSessionFromWorkingPath(path string, isActive bool) Session {
	session := Session{
		Name:           filepath.Base(path),
		WorkingPath:    path,
		RepositoryPath: path,
		Branch:         "",
		IsActive:       isActive,
	}

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
	MergeSessions(currentSessions []Session, newSessions []Session) []Session
}

func defaultMergeSessions(currentSessions []Session, newSessions []Session) []Session {
	return append(currentSessions, newSessions...)
}

func FindSessions(sources []SessionFinder) ([]Session, error) {
	var sessions []Session
	for _, source := range sources {
		sourceSessions, err := source.FindSessions()
		if err != nil {
			return nil, err
		}
		sessions = source.MergeSessions(sessions, sourceSessions)
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

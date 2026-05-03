package session

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func findSessionsFromPath(path string) ([]Session, error) {
	path, err := expandPathHomeDir(path)
	if err != nil {
		return nil, err
	}
	realPaths, err := expandPathWildcards(path)
	if err != nil {
		return nil, err
	}
	sessions := make([]Session, 0)
	for _, path := range realPaths {
		session, err := sessionChildrenOfRealPath(path)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session...)
	}
	return sessions, nil
}

func expandPathHomeDir(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("error when expanding path home dir %s: %w", path, err)
		}
		return filepath.Join(homeDir, path[1:]), nil
	}
	return path, nil
}

func expandPathWildcards(path string) ([]string, error) {
	pathParts := strings.Split(path, string(filepath.Separator))

	currentPath := ""

	realPaths := make([]string, 0)

	for i, part := range pathParts {
		if part == "*" {
			partChildren, err := os.ReadDir(currentPath)
			if errors.Is(err, fs.ErrNotExist) {
				return realPaths, nil // non existant directories can be safely ignored
			}
			if err != nil {
				return nil, fmt.Errorf("error when expanding path wildcards %s: %w", path, err)
			}
			for _, child := range partChildren {
				if isVisibleDirectory(child) {
					newPath := filepath.Join(currentPath, child.Name())
					if i < len(pathParts)-1 {
						newPath = filepath.Join(newPath, filepath.Join(pathParts[i+1:]...))
					}
					childPaths, err := expandPathWildcards(newPath)
					if err != nil {
						return nil, err
					}
					realPaths = append(realPaths, childPaths...)
				}
			}
			return realPaths, nil
		}
		if i == 0 {
			currentPath = part
		} else {
			currentPath = currentPath + string(filepath.Separator) + part
		}
	}

	realPaths = append(realPaths, currentPath)

	return realPaths, nil
}

func sessionChildrenOfRealPath(realPath string) ([]Session, error) {
	children, err := os.ReadDir(realPath)
	if err != nil {
		return nil, fmt.Errorf("error when getting sessions of %s: %w", realPath, err)
	}

	sessions := make([]Session, 0)

	for _, child := range children {
		if isVisibleDirectory(child) {
			childPath := filepath.Join(realPath, child.Name())
			worktrees, err := findWorktreesFromRealPath(childPath)

			// There are no worktrees, or there is only one worktree in the same or a parent directory
			if err != nil ||
				len(worktrees) == 0 ||
				(len(worktrees) == 1 &&
					(worktrees[0].Path == childPath || len(worktrees[0].Path) < len(childPath))) {
				branch := ""
				if len(worktrees) == 1 {
					branch = worktrees[0].Branch
				}
				sessions = append(sessions, Session{
					Name:        child.Name(),
					Path:        childPath,
					ProjectPath: childPath,
					Branch:      branch,
					IsActive:    false,
				})
			} else {
				for _, worktree := range worktrees {
					sessions = append(sessions, Session{
						Name:        child.Name() + "[" + filepath.Base(worktree.Path) + "]",
						Path:        worktree.Path,
						ProjectPath: childPath,
						Branch:      worktree.Branch,
						IsActive:    false,
					})
				}
			}
		}
	}

	return sessions, nil
}

func isVisibleDirectory(d fs.DirEntry) bool {
	return d.IsDir() && d.Name()[0] != '.'
}

package session

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type PathSessionFinder struct {
	SearchPaths  []string
	IncludePaths []string
}

func (f PathSessionFinder) FindSessions() ([]Session, error) {
	var sessions []Session
	for _, search := range f.SearchPaths {
		paths, err := searchPath(search)
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			sessions = append(sessions, NewSessionFromWorkingPath(path))
		}
	}
	for _, path := range f.IncludePaths {
		path, err := expandPathHomeDir(path)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, NewSessionFromWorkingPath(path))
	}
	return sessions, nil
}

func searchPath(path string) ([]string, error) {
	path, err := expandPathHomeDir(path)
	if err != nil {
		return nil, err
	}
	realPaths, err := expandPathWildcards(path)
	if err != nil {
		return nil, err
	}
	foundPaths := make([]string, 0)
	for _, path := range realPaths {
		children, err := childrenOfPath(path)
		if err != nil {
			return nil, err
		}
		foundPaths = append(foundPaths, children...)
	}
	return foundPaths, nil
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

func childrenOfPath(path string) ([]string, error) {
	foundPaths := make([]string, 0)
	children, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("error when getting children of %s: %w", path, err)
	}
	for _, child := range children {
		if isVisibleDirectory(child) {
			foundPaths = append(foundPaths, filepath.Join(path, child.Name()))
		}
	}
	return foundPaths, nil
}

func isVisibleDirectory(d fs.DirEntry) bool {
	return d.IsDir() && d.Name()[0] != '.'
}

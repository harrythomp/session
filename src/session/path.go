package session

import (
	"errors"
	"io/fs"
	"os"
	"strings"
)

func findSessionsFromPath(path string) ([]Session, error) {
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

func expandPathWildcards(path string) ([]string, error) {
	pathParts := strings.Split(path, "/")

	currentPath := ""

	realPaths := make([]string, 0)

	for i, part := range pathParts {
		if part == "*" {
			partChildren, err := os.ReadDir(currentPath)
			if errors.Is(err, fs.ErrNotExist) {
				return realPaths, nil
			}
			if err != nil {
				return nil, err
			}
			for _, child := range partChildren {
				if isVisibleDirectory(child) {
					newPath := currentPath + "/" + child.Name()
					if i < len(pathParts)-1 {
						newPath = newPath + "/" + strings.Join(pathParts[i+1:], "/")
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
			currentPath = currentPath + "/" + part
		}
	}

	realPaths = append(realPaths, currentPath)

	return realPaths, nil
}

func sessionChildrenOfRealPath(realPath string) ([]Session, error) {
	children, err := os.ReadDir(realPath)
	if errors.Is(err, fs.ErrNotExist) {
		return make([]Session, 0), nil
	}
	if err != nil {
		return nil, err
	}

	sessions := make([]Session, 0)

	for _, child := range children {
		if isVisibleDirectory(child) {
			sessions = append(sessions, Session{
				Name:     child.Name(),
				Path:     realPath + "/" + child.Name(),
				IsActive: false,
			})
		}
	}

	return sessions, nil
}

func isVisibleDirectory(d fs.DirEntry) bool {
	return d.IsDir() && d.Name()[0] != '.'
}

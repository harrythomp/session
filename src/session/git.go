package session

import (
	"os/exec"
	"path/filepath"
	"strings"
)

type Worktree struct {
	Path   string
	Branch string
}

func findWorktreesFromRealPath(path string) ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = path
	outBytes, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	out := string(outBytes)

	worktrees := make([]Worktree, 0)

	currentTree := Worktree{}

	lines := strings.SplitSeq(out, "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "worktree") {
			currentTree.Path = line[len("worktree "):]
		} else if strings.HasPrefix(line, "branch") {
			currentTree.Branch = filepath.Base(line[len("branch "):])
		} else if line == "bare" {
			currentTree = Worktree{}
		} else if currentTree != (Worktree{}) && line == "" {
			worktrees = append(worktrees, currentTree)
			currentTree = Worktree{}
		}
	}

	return worktrees, nil
}

func getBranchFromRealPath(realPath string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = realPath
	outBytes, err := cmd.Output()
	if err != nil {
		return "", err
	}
	out := string(outBytes)
	return strings.TrimSpace(out), nil
}

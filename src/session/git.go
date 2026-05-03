package session

import (
	"fmt"
	"os/exec"
	"strings"
)

type Worktree struct {
	Path   string
	Branch string
}

func findWorktreesFromRealPath(realPath string) ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = realPath
	outBytes, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error when getting git worktree list for %s: %w", realPath, err)
	}
	out := string(outBytes)

	worktrees := make([]Worktree, 0)

	currentTree := Worktree{}

	lines := strings.SplitSeq(out, "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "worktree "); ok {
			currentTree.Path = after
		} else if after, ok := strings.CutPrefix(line, "branch refs/heads/"); ok {
			currentTree.Branch = after
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
		return "", fmt.Errorf("error when getting branch for %s: %w", realPath, err)
	}
	out := string(outBytes)
	return strings.TrimSpace(out), nil
}

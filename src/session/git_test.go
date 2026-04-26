package session

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_findWorktreesFromRealPath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs(cwd + "/../../")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		path    string
		want    []Worktree
		wantErr bool
	}{
		{
			name: "zero-depth-worktree",
			path: root + "/test/zero-depth-worktree/project-one",
			want: []Worktree{
				{
					Path:   root + "/test/zero-depth-worktree/project-one/main",
					Branch: "main",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := findWorktreesFromRealPath(tt.path)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("findWorktreesFromPath() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("findWorktreesFromPath() succeeded unexpectedly")
			}
			if len(got) != len(tt.want) {
				t.Fatalf("findWorktreesFromPath() returned %d worktrees, want %d", len(got), len(tt.want))
			}
			for i, want := range tt.want {
				if got[i] != want {
					t.Errorf("findWorktreesFromPath()[%d] = %v, want %v", i, got[i], want)
				}
			}
		})
	}
}

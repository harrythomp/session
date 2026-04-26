package session

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_findSessionsFromPath(t *testing.T) {
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
		want    []Session
		wantErr bool
	}{
		{
			name: "zero-depth-basic",
			path: root + "/test/zero-depth-basic",
			want: []Session{
				{
					Name: "project-one",
					Path: root + "/test/zero-depth-basic/project-one",
				},
				{
					Name: "project-three",
					Path: root + "/test/zero-depth-basic/project-three",
				},
				{
					Name: "project-two",
					Path: root + "/test/zero-depth-basic/project-two",
				},
			},
			wantErr: false,
		},
		{
			name: "one-depth-basic",
			path: root + "/test/one-depth-basic/*",
			want: []Session{
				{
					Name: "deep-project-one",
					Path: root + "/test/one-depth-basic/skip/deep-project-one",
				},
				{
					Name: "deep-project-two",
					Path: root + "/test/one-depth-basic/skip/deep-project-two",
				},
			},
			wantErr: false,
		},
		{
			name: "zero-depth-worktree",
			path: root + "/test/zero-depth-worktree",
			want: []Session{
				{
					Name: "project-one[main]",
					Path: root + "/test/zero-depth-worktree/project-one/main",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := findSessionsFromPath(tt.path)
			if gotErr != nil || got == nil {
				if !tt.wantErr {
					t.Errorf("FindSessions() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("FindSessions() succeeded unexpectedly")
			}
			if len(got) != len(tt.want) {
				t.Fatalf("FindSessions() returned %d sessions, want %d", len(got), len(tt.want))
			}
			for i, want := range tt.want {
				if got[i].Name != want.Name || got[i].Path != want.Path {
					t.Errorf("FindSessions()[%d] = %v, want %v", i, got[i], want)
				}
			}
		})
	}
}

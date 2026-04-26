package session

import (
	"harry/session/src/config"
	"os"
	"path/filepath"
	"testing"
)

func Test_FindSessions(t *testing.T) {
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
		conf    config.Config
		want    []Session
		wantErr bool
	}{
		{
			name: "basic",
			conf: config.Config{
				SearchPaths: []string{
					root + "/test/zero-depth-basic",
					root + "/test/one-depth-basic/*",
				},
				IncludePaths: []string{
					root + "/test/extra-project",
				},
			},
			want: []Session{
				{
					Name:     "session",
					Path:     root,
					IsActive: true,
				},
				{
					Name:     "extra-project",
					Path:     root + "/test/extra-project",
					IsActive: false,
				},
				{
					Name:     "deep-project-one",
					Path:     root + "/test/one-depth-basic/skip/deep-project-one",
					IsActive: false,
				},
				{
					Name:     "deep-project-two",
					Path:     root + "/test/one-depth-basic/skip/deep-project-two",
					IsActive: false,
				},
				{
					Name:     "project-one",
					Path:     root + "/test/zero-depth-basic/project-one",
					IsActive: false,
				},
				{
					Name:     "project-three",
					Path:     root + "/test/zero-depth-basic/project-three",
					IsActive: false,
				},
				{
					Name:     "project-two",
					Path:     root + "/test/zero-depth-basic/project-two",
					IsActive: false,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := FindSessions(tt.conf)
			if gotErr != nil {
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

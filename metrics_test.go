package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestReadMetrics(t *testing.T) {
	for _, tc := range []struct {
		name      string
		content   string
		wantCount int
		wantErr   error
	}{
		{
			name: "empty",
		},
		{
			name:    "bad format",
			content: "hello<world>{ nope",
			wantErr: cmpopts.AnyError,
		},
		{
			name:      "valid",
			content:   "foo 1\nbar 2\n",
			wantCount: 2,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "metrics")

			if err := os.WriteFile(path, []byte(tc.content), 0o644); err != nil {
				t.Fatalf("Writing metrics failed: %v", err)
			}

			got, err := readMetrics(path)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}

			if err == nil {
				if diff := cmp.Diff(tc.wantCount, len(got)); diff != "" {
					t.Errorf("Count diff (-want +got):\n%s", diff)
				}
			}
		})
	}
}

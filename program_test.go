package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func disableLogOutput(t *testing.T) {
	t.Helper()

	previous := log.Writer()

	log.SetOutput(io.Discard)

	t.Cleanup(func() {
		log.SetOutput(previous)
	})
}

func TestProgramOptionsRegisterFlags(t *testing.T) {
	for _, tc := range []struct {
		name    string
		environ map[string]string
		args    []string
		want    programOptions
		wantErr error
	}{
		{
			name: "empty",
			want: programOptions{
				Retries:    2,
				RetryDelay: 10 * time.Second,
			},
		},
		{
			name: "args only",
			args: []string{
				"-gateway=http://localhost:1234",
				"-retries=100",
				"-retry_delay=2h",
				"-metrics=/a/b/c",
			},
			want: programOptions{
				URL:         "http://localhost:1234",
				Retries:     100,
				RetryDelay:  2 * time.Hour,
				MetricsFile: "/a/b/c",
			},
		},
		{
			name: "env only",
			environ: map[string]string{
				"PROMPUSH_URL":          "http://localhost:9876",
				"PROMPUSH_RETRIES":      "123",
				"PROMPUSH_RETRY_DELAY":  "4h",
				"PROMPUSH_METRICS_FILE": "/not/found",
			},
			want: programOptions{
				URL:         "http://localhost:9876",
				Retries:     123,
				RetryDelay:  4 * time.Hour,
				MetricsFile: "/not/found",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.environ == nil {
				tc.environ = map[string]string{}
			}

			var o programOptions

			fs := flag.NewFlagSet(tc.name, flag.ContinueOnError)
			fs.Usage = func() {}

			err := o.registerFlags(fs, tc.environ)

			if diff := cmp.Diff(tc.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Error diff (-want +got):\n%s", diff)
			}

			if err == nil {
				if err := fs.Parse(tc.args); err != nil {
					t.Fatalf("Parsing %q failed: %v", tc.args, err)
				}

				if diff := cmp.Diff(tc.want, o, cmpopts.EquateEmpty()); diff != "" {
					t.Errorf("Options diff (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestProgramRun(t *testing.T) {
	disableLogOutput(t)

	count := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if count == 0 {
			http.Error(w, "", http.StatusTeapot)
		}

		count++
	}))
	t.Cleanup(ts.Close)

	p, err := newProgram(programOptions{
		URL:         ts.URL,
		Job:         t.Name(),
		MetricsFile: os.DevNull,
		Retries:     2,
	})
	if err != nil {
		t.Errorf("newProgram() failed: %v", err)
	}

	if err == nil {
		err = p.run(t.Context())
		if err != nil {
			t.Errorf("run() failed: %v", err)
		}
	}
}

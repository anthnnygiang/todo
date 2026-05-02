package main

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/kong"
)

var (
	testDataDir = "testdata"
	lsEmptyFile = "ls_empty.txt"
	lsTwoFile   = "ls_two.txt"
	addOneFile  = "add_one.txt"
	rmTwoFile   = "rm_two.txt"
	rmAllFile   = "rm_all.txt"
)

func TestCLICmds(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		file   string
		output string // user-facing stdout
		want   string // persisted file content
	}{
		{
			name:   "LsEmpty",
			args:   []string{"ls"},
			file:   lsEmptyFile,
			output: "\x1b[32mall done!\x1b[0m\n",
			want:   "",
		},
		{
			name: "LsTwo",
			args: []string{"ls"},
			file: lsTwoFile,
			output: "\x1b[32m\x1b[1m1.\x1b[0m first todo\n" +
				"\x1b[32m\x1b[1m2.\x1b[0m second todo\n",
			want: "first todo\nsecond todo\n",
		},
		{
			name:   "AddOne",
			args:   []string{"add", "new todo"},
			file:   addOneFile,
			output: "\x1b[32m\x1b[1m1.\x1b[0m new todo\n",
			want:   "new todo\n",
		},
		{
			name: "RmTwo",
			args: []string{"done", "1", "2"},
			file: rmTwoFile,
			output: "\x1b[32m\x1b[1m1.\x1b[0m third todo\n" +
				"\x1b[32m\x1b[1m2.\x1b[0m fourth todo\n",
			want: "third todo\nfourth todo\n",
		},
		{
			name:   "RmAll",
			args:   []string{"done"},
			file:   rmAllFile,
			output: "\x1b[32mall done!\x1b[0m\n",
			want:   "",
		},
	}

	// prepare fresh testdata
	tmp := t.TempDir()
	if err := copyDir(testDataDir, tmp); err != nil {
		t.Fatalf("copyDir(%q, tmp) failed: %v", testDataDir, err)
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel() // each test has it's own file

			// open the specific todo file
			fpath := filepath.Join(tmp, tc.file)
			file, err := os.OpenFile(fpath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
			if err != nil {
				t.Fatalf("openFile(%q) failed: %v", tc.file, err)
			}
			t.Cleanup(func() {
				if err := file.Close(); err != nil {
					t.Errorf("close(%q) failed: %v", tc.file, err)
				}
			})

			// use buffer instead of os.Stdout for testing
			out := &bytes.Buffer{}
			spec := CLI{
				Ls:   LsCmd{Out: out, File: file},
				Add:  AddCmd{Out: out, File: file},
				Done: DoneCmd{Out: out, File: file},
			}

			parser := kong.Must(&spec)
			ctx, err := parser.Parse(tc.args)
			if err != nil {
				t.Fatalf("parse(%v) failed: %v", tc.args, err)
			}
			if err := ctx.Run(); err != nil {
				t.Fatalf("run failed: %v", err)
			}

			got := out.String()
			if got != tc.output {
				t.Errorf("expected output %q, got %q", tc.output, got)
			}

			fileBytes, err := os.ReadFile(fpath)
			if err != nil {
				t.Fatalf("readFile(%q) failed: %v", tc.file, err)
			}

			gotFile := string(fileBytes)
			if gotFile != tc.want {
				t.Errorf("expected file contents %q, got %q", tc.want, gotFile)
			}
		})
	}
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0700)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}

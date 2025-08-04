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

func TestCLICmds(t *testing.T) {
	tests := []struct {
		name string
		args []string
		file string
		want string
	}{
		{
			name: "LsEmpty",
			args: []string{"ls"},
			file: lsEmptyFile,
			want: "\x1b[32mall done!\x1b[0m\n",
		},
		{
			name: "LsTwo",
			args: []string{"ls"},
			file: lsTwoFile,
			want: "\x1b[32m\x1b[1m1.\x1b[0m first todo\n" +
				"\x1b[32m\x1b[1m2.\x1b[0m second todo\n",
		},
		{
			name: "AddOne",
			args: []string{"add", "new todo"},
			file: addOneFile,
			want: "\x1b[32m\x1b[1m1.\x1b[0m new todo\n",
		},
		{
			name: "RmTwo",
			args: []string{"rm", "1", "2"},
			file: rmTwoFile,
			want: "\x1b[32m\x1b[1m1.\x1b[0m third todo\n" +
				"\x1b[32m\x1b[1m2.\x1b[0m fourth todo\n",
		},
		{
			name: "RmAll",
			args: []string{"rm"},
			file: rmAllFile,
			want: "\x1b[32mall done!\x1b[0m\n",
		},
	}

	// Prepare fresh testdata
	tmp := t.TempDir()
	if err := copyDir(testDataDir, tmp); err != nil {
		t.Fatalf("copyDir failed: %v", err)
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel() // each test has it's own file

			// open the specific todo file
			fpath := filepath.Join(tmp, tc.file)
			file, err := openFile(fpath)
			if err != nil {
				t.Fatalf("openFile(%q) failed: %v", tc.file, err)
			}

			// use buffer instead of os.Stdout for testing
			out := &bytes.Buffer{}
			spec := CLI{
				Ls:  LsCmd{Out: out, File: file},
				Add: AddCmd{Out: out, File: file},
				Rm:  RmCmd{Out: out, File: file},
			}

			parser := kong.Must(&spec)
			ctx, err := parser.Parse(tc.args)
			if err != nil {
				t.Fatalf("Parse(%v) failed: %v", tc.args, err)
			}
			if err := ctx.Run(ctx); err != nil {
				t.Fatalf("Run failed: %v", err)
			}

			got := out.String()
			if got != tc.want {
				t.Errorf("expected output %q, got %q", tc.want, got)
			}
		})
	}
}

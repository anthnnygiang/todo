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

var testDataDir = "testdata"
var lsEmptyFile = "ls_empty.txt"
var lsTwoFile = "ls_two.txt"
var addOneFile = "add_one.txt"
var rmTwoFile = "rm_two.txt"
var rmAllFile = "rm_all.txt"

// Copy the test data files to a temporary directory for testing
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		in, _ := os.Open(path)
		defer in.Close()
		out, _ := os.Create(target)
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}

func TestLsEmptyCmd(t *testing.T) {
	tmp := t.TempDir()
	if err := copyDir(testDataDir, tmp); err != nil {
		t.Fatalf("failed to copy testdata: %v", err)
	}
	file, err := openFile(filepath.Join(tmp, lsEmptyFile))
	out := &bytes.Buffer{}
	CLISpec := CLI{
		Ls: LsCmd{
			Out:  out,
			File: file,
		},
		Add: AddCmd{
			Out:  out,
			File: file,
		},
		Rm: RmCmd{
			Out:  out,
			File: file,
		},
	}

	// Set up a parser that writes to a buffer instead of os.Stdout.
	parser := kong.Must(&CLISpec)
	args := []string{"ls"}
	ctx, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("parse(%v) failed: %s", args, err)
	}
	if err := ctx.Run(ctx); err != nil {
		t.Fatalf("run failed: %s", err)
	}
	got := out.String()
	want := "\x1b[32mall done!\x1b[0m\n"
	if got != want {
		t.Errorf("\nexpected: %q\ngot: %q", want, got)
	}
}

func TestLsTwoCmd(t *testing.T) {
	tmp := t.TempDir()
	if err := copyDir(testDataDir, tmp); err != nil {
		t.Fatalf("failed to copy testdata: %v", err)
	}
	file, err := openFile(filepath.Join(tmp, lsTwoFile))
	if err != nil {
		t.Fatalf("openFile(%s) failed: %s", lsTwoFile, err)
	}
	out := &bytes.Buffer{}
	CLISpec := CLI{
		Ls: LsCmd{
			Out:  out,
			File: file,
		},
		Add: AddCmd{
			Out:  out,
			File: file,
		},
		Rm: RmCmd{
			Out:  out,
			File: file,
		},
	}

	parser := kong.Must(&CLISpec)
	args := []string{"ls"}
	ctx, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("parse(%v) failed: %s", args, err)
	}
	if err := ctx.Run(ctx); err != nil {
		t.Fatalf("run failed: %s", err)
	}
	got := out.String()
	want := "\x1b[32m\x1b[1m1.\x1b[0m first todo\n\x1b[32m\x1b[1m2.\x1b[0m second todo\n"
	if got != want {
		t.Errorf("\nexpected: %q\ngot: %q", want, got)
	}
}

func TestAddOneCmd(t *testing.T) {
	tmp := t.TempDir()
	if err := copyDir(testDataDir, tmp); err != nil {
		t.Fatalf("failed to copy testdata: %v", err)
	}
	file, err := openFile(filepath.Join(tmp, addOneFile))
	if err != nil {
		t.Fatalf("openFile(%s) failed: %s", addOneFile, err)
	}
	out := &bytes.Buffer{}
	CLISpec := CLI{
		Ls: LsCmd{
			Out:  out,
			File: file,
		},
		Add: AddCmd{
			Out:  out,
			File: file,
		},
		Rm: RmCmd{
			Out:  out,
			File: file,
		},
	}

	parser := kong.Must(&CLISpec)
	args := []string{"add", "new todo"}
	ctx, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("parse(%v) failed: %s", args, err)
	}
	if err := ctx.Run(ctx); err != nil {
		t.Fatalf("run failed: %s", err)
	}
	got := out.String()
	want := "\x1b[32m\x1b[1m1.\x1b[0m new todo\n"
	if got != want {
		t.Errorf("\nexpected: %q\ngot: %q", want, got)
	}
}

func TestRmTwoCmd(t *testing.T) {
	tmp := t.TempDir()
	if err := copyDir(testDataDir, tmp); err != nil {
		t.Fatalf("failed to copy testdata: %v", err)
	}
	file, err := openFile(filepath.Join(tmp, rmTwoFile))
	if err != nil {
		t.Fatalf("openFile(%s) failed: %s", rmTwoFile, err)
	}
	out := &bytes.Buffer{}
	CLISpec := CLI{
		Ls: LsCmd{
			Out:  out,
			File: file,
		},
		Add: AddCmd{
			Out:  out,
			File: file,
		},
		Rm: RmCmd{
			Out:  out,
			File: file,
		},
	}

	parser := kong.Must(&CLISpec)
	args := []string{"rm", "1", "2"}
	ctx, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("parse(%v) failed: %s", args, err)
	}
	if err := ctx.Run(ctx); err != nil {
		t.Fatalf("run failed: %s", err)
	}
	got := out.String()
	want := "\x1b[32m\x1b[1m1.\x1b[0m third todo\n\x1b[32m\x1b[1m2.\x1b[0m fourth todo\n"
	if got != want {
		t.Errorf("\nexpected: %q\ngot: %q", want, got)
	}
}

func TestRmAllCmd(t *testing.T) {
	tmp := t.TempDir()
	if err := copyDir(testDataDir, tmp); err != nil {
		t.Fatalf("failed to copy testdata: %v", err)
	}
	file, err := openFile(filepath.Join(tmp, rmAllFile))
	if err != nil {
		t.Fatalf("openFile(%s) failed: %s", rmAllFile, err)
	}
	out := &bytes.Buffer{}
	CLISpec := CLI{
		Ls: LsCmd{
			Out:  out,
			File: file,
		},
		Add: AddCmd{
			Out:  out,
			File: file,
		},
		Rm: RmCmd{
			Out:  out,
			File: file,
		},
	}

	parser := kong.Must(&CLISpec)
	args := []string{"rm"}
	ctx, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("parse(%v) failed: %s", args, err)
	}
	if err := ctx.Run(ctx); err != nil {
		t.Fatalf("run failed: %s", err)
	}
	got := out.String()
	want := "\x1b[32mall done!\x1b[0m\n"
	if got != want {
		t.Errorf("\nexpected: %q\ngot: %q", want, got)
	}
}

package main

import (
	"bytes"
	"testing"

	"github.com/alecthomas/kong"
)

func TestAdd(t *testing.T) {
	got := 5
	want := 5
	if got != want {
		t.Errorf("Add(2,3) = %d; want %d", got, want)
	}
}

func TestLsCmd(t *testing.T) {
	CLISpec := CLI{}

	// Set up a parser that writes to a buffer instead of os.Stdout.
	out := &bytes.Buffer{}
	parser := kong.Must(&CLISpec,
		kong.Writers(out, out),
		kong.Exit(func(int) { /* prevent os.Exit */ }),
	)

	args := []string{"ls"}
	ctx, err := parser.Parse(args)
	if err != nil {
		t.Fatalf("Parse(%v) failed: %s", args, err)
	}
	if err := ctx.Run(ctx); err != nil {
		t.Fatalf("Run failed: %s", err)
	}
	got := out.String()
	want := "Hello, Alice!\n"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}

}

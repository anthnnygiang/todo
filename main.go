package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
)

const Reset = "\033[0m"
const Bold = "\033[1m"
const Red = "\033[31m"
const Green = "\033[32m"

const todoDirectory = ".todo"

// define top level CLI commands
type CLI struct {
	Project string `short:"p" default:"todo" help:"Project todo list to use."`

	Ls  LsCmd  `cmd:"" help:"List all todo items."`
	Add AddCmd `cmd:"" help:"Add a todo item."`
	Rm  RmCmd  `cmd:"" help:"Remove one or more todo items. If no numbers are provided, remove all todo items."`
}

// LsCmd lists all current todo items
type LsCmd struct {
	Out  io.Writer `kong:"-"`
	File *os.File  `kong:"-"`
}

// AddCmd adds a new todo item using the provided title
type AddCmd struct {
	Title string    `help:"Title of the todo item." arg:""`
	Out   io.Writer `kong:"-"`
	File  *os.File  `kong:"-"`
}

// RmCmd removes one or more todo items, or clears the list if no numbers are provided
type RmCmd struct {
	Number []string  `help:"Todo items to remove." arg:"" optional:""`
	Out    io.Writer `kong:"-"`
	File   *os.File  `kong:"-"`
}

func main() {
	// build the CLI commands with shared dependencies
	CLISpec := CLI{
		Ls: LsCmd{
			Out: os.Stdout,
		},
		Add: AddCmd{
			Out: os.Stdout,
		},
		Rm: RmCmd{
			Out: os.Stdout,
		},
	}

	// Parse command-line input before opening the todo file so flags can select it.
	cmd := kong.Parse(&CLISpec)

	// Resolve the todo file path in the user's home directory.
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	todoDir := filepath.Join(home, todoDirectory)

	project := CLISpec.Project
	if project != "todo" {
		invalidNames := []string{"", ".", ".."}
		if slices.Contains(invalidNames, project) {
			fmt.Fprintf(os.Stderr, "invalid project name: %s\n", project)
			os.Exit(1)
		}
	}
	todoFile := project + ".txt"
	fpath := filepath.Join(todoDir, todoFile)

	if err := os.MkdirAll(todoDir, 0700); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// create the file if it does not already exist
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		fmt.Printf("creating '%s' in '%s' directory.\n", todoFile, todoDirectory)
	}

	// open the todo file for reading and writing
	file, err := os.OpenFile(fpath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		os.Exit(1)
	}

	// close the file before exiting
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}(file)

	// set the project file for each command
	CLISpec.Ls.File = file
	CLISpec.Add.File = file
	CLISpec.Rm.File = file

	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Run LsCmd prints all todo items currently stored in the file.
func (c *LsCmd) Run() error {
	return list(c.Out, c.File)
}

// Run AddCmd appends the new todo item to the file and prints the updated list.
func (c *AddCmd) Run() error {
	if _, err := c.File.WriteString(c.Title + "\n"); err != nil {
		return err
	}
	return list(c.Out, c.File)
}

// Run RmCmd removes selected todo items or clears the file if no numbers are provided.
func (c *RmCmd) Run() error {
	// if no numbers are provided, remove all todo items
	if len(c.Number) == 0 {
		if err := clear(c.File); err != nil {
			return err
		}
		return list(c.Out, c.File)
	}

	// create a lookup table of item numbers to remove
	numbersMap := make(map[int]bool)
	for _, n := range c.Number {
		num, err := strconv.Atoi(n)
		if err != nil {
			return err
		}
		numbersMap[num] = true
	}

	// read the file and keep only the items that should remain
	var remainingTodos []string
	fileScanner := bufio.NewScanner(c.File)
	fileScanner.Split(bufio.ScanLines) // each scan returns one line of text
	for i := 1; fileScanner.Scan(); i++ {
		if numbersMap[i] {
			continue
		}
		remainingTodos = append(remainingTodos, fileScanner.Text())
	}
	if err := fileScanner.Err(); err != nil {
		return err
	}

	// clear the file and write the remaining todos back to it
	if err := clear(c.File); err != nil {
		return err
	}

	for _, line := range remainingTodos {
		_, err := c.File.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	return list(c.Out, c.File)
}

// list prints all items in the file
func list(out io.Writer, file *os.File) error {
	var filename = file.Name()
	fmt.Fprintf(out, "%s%s%s:%s\n", Green, Bold, strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename)), Reset)
	// reset the file pointer to the beginning before reading
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines) // each scan returns one line of text

	var todoLines []string
	for fileScanner.Scan() {
		todoLines = append(todoLines, fileScanner.Text())
	}
	if err := fileScanner.Err(); err != nil {
		return err
	}

	if len(todoLines) == 0 {
		fmt.Fprintf(out, "%sall done!%s\n", Green, Reset)
	}

	// print each todo item with a numbered index
	for i, line := range todoLines {
		fmt.Fprintf(out, "%s%s%d.%s %s\n", Green, Bold, i+1, Reset, line)
	}
	return nil
}

// clear removes all contents from the todo file and resets its cursor.
func clear(file *os.File) error {
	if err := file.Truncate(0); err != nil {
		return err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return nil
}

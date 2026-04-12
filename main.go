package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/alecthomas/kong"
)

const Reset = "\033[0m"
const Bold = "\033[1m"
const Red = "\033[31m"
const Green = "\033[32m"

// todosFile in home directory
var todosFile = ".todos.txt"

// define top level CLI commands
type CLI struct {
	Ls   LsCmd   `cmd:"" help:"List all todo items."`
	Add  AddCmd  `cmd:"" help:"Add a todo item."`
	Done DoneCmd `cmd:"" help:"Complete one or more todo items. If no numbers are provided, complete all todo items."`
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

// DoneCmd removes one or more todo items, or clears the list if no numbers are provided
type DoneCmd struct {
	Number []string  `help:"Todo items to complete." arg:"" optional:""`
	Out    io.Writer `kong:"-"`
	File   *os.File  `kong:"-"`
}

func main() {
	// Resolve the todo file path in the user's home directory.
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	fpath := filepath.Join(home, todosFile)

	// create the file if it does not already exist
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		fmt.Printf("%s'%s' does not exist.%s\n", Red, todosFile, Reset)
		fmt.Printf("%screating '%s' in home directory...%s\n", Red, todosFile, Reset)
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

	// build the CLI commands with shared dependencies
	CLISpec := CLI{
		Ls: LsCmd{
			Out:  os.Stdout,
			File: file,
		},
		Add: AddCmd{
			Out:  os.Stdout,
			File: file,
		},
		Done: DoneCmd{
			Out:  os.Stdout,
			File: file,
		},
	}

	// Parse command-line input and Run the selected command
	cmd := kong.Parse(&CLISpec)
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Run LsCmd prints all todo items currently stored in the file.
func (c *LsCmd) Run() error {
	list(c.Out, c.File)
	return nil
}

// Run AddCmd appends the new todo item to the file and prints the updated list.
func (c *AddCmd) Run() error {
	input := c.Title
	bytes := make([]byte, 0)
	if _, err := c.File.Write(fmt.Appendf(bytes, "%s\n", input)); err != nil {
		return err
	}
	list(c.Out, c.File)
	return nil
}

// Run DoneCmd removes selected todo items or clears the file if no numbers are provided.
func (c *DoneCmd) Run() error {
	// if no numbers are provided, complete all todo items
	if len(c.Number) == 0 {
		if err := clear(c.File); err != nil {
			return err
		}
		list(c.Out, c.File)
		return nil
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

	list(c.Out, c.File)
	return nil
}

// list prints all items in the file
func list(out io.Writer, file *os.File) error {
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

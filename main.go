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

var (
	repositoryPath = "dev/.zzz/todo"
	todosFile      = "todos.txt"
)

type CLI struct {
	Ls  LsCmd  `cmd:"" help:"List all todo items."`
	Add AddCmd `cmd:"" help:"Add a todo item."`
	Rm  RmCmd  `cmd:"" help:"Complete one or more todo items. If no numbers are provided, complete all todo items."`
}
type LsCmd struct {
	Out  io.Writer `kong:"-"`
	File *os.File  `kong:"-"`
}
type AddCmd struct {
	Title string    `help:"Title of the todo item." arg:""`
	Out   io.Writer `kong:"-"`
	File  *os.File  `kong:"-"`
}
type RmCmd struct {
	Number []string  `help:"Todo items to complete." arg:"" optional:""`
	Out    io.Writer `kong:"-"`
	File   *os.File  `kong:"-"`
}

func main() {
	// open Todo file
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	filename := filepath.Join(home, repositoryPath, todosFile)
	file, err := openFile(filename)
	if err != nil {
		return
	}
	// errors are handled automatically
	CLISpec := CLI{
		Ls: LsCmd{
			Out:  os.Stdout,
			File: file,
		},
		Add: AddCmd{
			Out:  os.Stdout,
			File: file,
		},
		Rm: RmCmd{
			Out:  os.Stdout,
			File: file,
		},
	}
	ctx := kong.Parse(&CLISpec)
	if err := ctx.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (c *LsCmd) Run() error {
	list(c.Out, c.File)
	defer closeFile(c.File)
	return nil
}

func (c *AddCmd) Run() error {
	input := c.Title
	bytes := make([]byte, 0)
	if _, err := c.File.Write(fmt.Appendf(bytes, "%s\n", input)); err != nil {
		return err
	}
	list(c.Out, c.File)
	defer closeFile(c.File)
	return nil
}

func (c *RmCmd) Run() error {
	// if no numbers are provided, complete all todo items
	if len(c.Number) == 0 {
		c.File.Truncate(0)
		list(c.Out, c.File)
		return nil
	}

	// create a map of numbers
	numbersMap := make(map[int]bool)
	for _, n := range c.Number {
		num, err := strconv.Atoi(n)
		if err != nil {
			return err
		}
		numbersMap[num] = true
	}

	// read the file and keep todos that are not in numbersMap
	var remainingTodos []string
	fileScanner := bufio.NewScanner(c.File)
	fileScanner.Split(bufio.ScanLines)
	for i := 1; fileScanner.Scan(); i++ {
		if numbersMap[i] {
			continue
		}
		remainingTodos = append(remainingTodos, fileScanner.Text())
	}

	if err := c.File.Truncate(0); err != nil {
		return err
	}
	if _, err := c.File.Seek(0, io.SeekStart); err != nil {
		return err
	}

	for _, line := range remainingTodos {
		bytes := make([]byte, 0)
		_, err := c.File.Write(fmt.Appendf(bytes, "%s\n", line))
		if err != nil {
			return err
		}
	}
	list(c.Out, c.File)
	defer closeFile(c.File)
	return nil
}

// list prints all items in the file.
func list(out io.Writer, file *os.File) error {
	// Reset the file pointer to the beginning.
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	var todoLines []string
	for fileScanner.Scan() {
		todoLines = append(todoLines, fileScanner.Text())
	}
	if len(todoLines) == 0 {
		fmt.Fprintf(out, "%sall done!%s\n", Green, Reset)
	}
	for i, line := range todoLines {
		fmt.Fprintf(out, "%s%s%d.%s %s\n", Green, Bold, i+1, Reset, line)
	}
	return nil
}

func openFile(filename string) (*os.File, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("%s%s does not exist.%s\n", Red, todosFile, Reset)
		fmt.Printf("%screating %s...%s\n", Red, todosFile, Reset)
	}
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func closeFile(file *os.File) error {
	err := file.Close()
	if err != nil {
		return err
	}
	return nil
}

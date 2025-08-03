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
const Black = "\033[30m"
const Red = "\033[31m"
const Green = "\033[32m"
const Yellow = "\033[33m"
const Blue = "\033[34m"
const Magenta = "\033[35m"
const Cyan = "\033[36m"
const Gray = "\033[37m"
const BrightBlack = "\033[90m"
const BrightRed = "\033[91m"
const BrightGreen = "\033[92m"
const BrightYellow = "\033[93m"
const BrightBlue = "\033[94m"
const BrightMagenta = "\033[95m"
const BrightCyan = "\033[96m"
const White = "\033[97m"

const Bold = "\033[1m"
const Dim = "\033[2m"
const Italic = "\033[3m"
const Underline = "\033[4m"
const Blink = "\033[5m"
const Reverse = "\033[7m"
const Hidden = "\033[8m"
const StrikeThrough = "\033[9m"

var repositoryPath = "dev/.zzz/todo"
var todosFile = "todos.txt"
var todosTestFile = "todos_test.txt"

type CLI struct {
	Ls  LsCmd  `cmd:"" help:"List all todo items."`
	Add AddCmd `cmd:"" help:"Add a todo item."`
	Rm  RmCmd  `cmd:"" help:"Complete one or more todo items. If no numbers are provided, complete all todo items."`
}
type LsCmd struct {
	Out io.Writer `kong:"-"`
}
type AddCmd struct {
	Title string    `help:"Title of the todo item." arg:""`
	Out   io.Writer `kong:"-"`
}
type RmCmd struct {
	Number []string  `help:"Todo items to complete." arg:"" optional:""`
	Out    io.Writer `kong:"-"`
}

func main() {
	// errors are handled automatically
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
	ctx := kong.Parse(&CLISpec)
	if err := ctx.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func (c *LsCmd) Run() error {
	file, err := openFile()
	if err != nil {
		return err
	}
	list(c.Out, file)
	defer closeFile(file)
	return nil
}

func (c *AddCmd) Run() error {
	file, err := openFile()
	if err != nil {
		return err
	}

	input := c.Title
	bytes := make([]byte, 0)
	if _, err = file.Write(fmt.Appendf(bytes, "%s\n", input)); err != nil {
		return err
	}
	list(c.Out, file)
	defer closeFile(file)
	return nil
}

func (c *RmCmd) Run() error {
	file, err := openFile()
	if err != nil {
		return err
	}
	// if no numbers are provided, complete all todo items
	if len(c.Number) == 0 {
		file.Truncate(0)
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
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	for i := 1; fileScanner.Scan(); i++ {
		if numbersMap[i] {
			continue
		}
		remainingTodos = append(remainingTodos, fileScanner.Text())
	}

	if err = file.Truncate(0); err != nil {
		return err
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	for _, line := range remainingTodos {
		bytes := make([]byte, 0)
		_, err := file.Write(fmt.Appendf(bytes, "%s\n", line))
		if err != nil {
			return err
		}
	}
	list(c.Out, file)
	defer closeFile(file)
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

func openFile() (*os.File, error) {
	// open Todo file
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	filename := filepath.Join(home, repositoryPath, todosFile)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("%s%s does not exist.%s\n", Red, todosFile, Reset)
		fmt.Printf("%screating %s...%s\n", BrightYellow, todosFile, Reset)
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

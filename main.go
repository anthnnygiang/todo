package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/alecthomas/kong"
)

const Reset = "\033[0m"
const Red = "\033[31m"
const Green = "\033[32m"
const Yellow = "\033[33m"
const Blue = "\033[34m"
const Magenta = "\033[35m"
const Cyan = "\033[36m"
const Gray = "\033[37m"
const White = "\033[97m"

var CLI struct {
	Ls struct{} `cmd:"" help:"List all todo items."`

	Add struct {
		Dir   bool   `short:"d" help:"Add a todo directory."`
		Title string `arg:"" help:"Title of the todo item." `
	} `cmd:"" help:"Add a todo item."`

	Rm struct {
		Number []string `arg:"" optional:"" help:"Todo items to complete."`
	} `cmd:"" help:"Complete one or more todo items. If no numbers are provided, complete all todo items."`
}

var repositoryPath = "dev/.asleep/todo"
var defaultDir = "todos.txt"

func main() {
	// Open the file for reading and writing.
	home, err := os.UserHomeDir()
	check(err)
	filename := filepath.Join(home, repositoryPath, defaultDir)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("%s%s does not exist.%s\n", Yellow, defaultDir, Reset)
		fmt.Printf("%sCreating %s...%s\n", Yellow, defaultDir, Reset)
	}
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	check(err)
	defer func(file *os.File) {
		err := file.Close()
		check(err)
	}(file)

	// errors are handled automatically
	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "ls":
		list(file)

	case "add <title>":
		title := CLI.Add.Title
		isDir := CLI.Add.Dir
		if !isDir {
			bytes := make([]byte, 0)
			_, err = file.Write(fmt.Appendf(bytes, "%s\n", title))
			check(err)
			list(file)
			return
		}
		// create a empty text file
		filePath := filepath.Join(home, repositoryPath, title+".txt")
		if _, err := os.Stat(filePath); !os.IsNotExist(err) {
			fmt.Printf("%s%s already exists.%s\n", Red, title+".txt", Reset)
			return
		}
		file, err = os.OpenFile(filePath, os.O_CREATE, 0644)
		check(err)
		defer func(file *os.File) {
			err := file.Close()
			check(err)
		}(file)

	case "rm":
		err := os.Truncate(filename, 0)
		check(err)

	case "rm <number>":
		// Create a map of numbers to check quickly.
		numbersMap := make(map[int]bool)
		for _, n := range CLI.Rm.Number {
			num, err := strconv.Atoi(n)
			check(err)
			numbersMap[num] = true
		}

		// Read the file and skip lines that are in numbersMap.
		var remainingTodos []string
		fileScanner := bufio.NewScanner(file)
		fileScanner.Split(bufio.ScanLines)
		for i := 1; fileScanner.Scan(); i++ {
			if numbersMap[i] {
				continue
			}
			remainingTodos = append(remainingTodos, fileScanner.Text())
		}

		err = file.Truncate(0)
		check(err)

		for _, line := range remainingTodos {
			bytes := make([]byte, 0)
			_, err := file.Write(fmt.Appendf(bytes, "%s\n", line))
			check(err)
		}
		list(file)

	default:
		fmt.Printf("%sUnknown command: %s%s\n", Red, ctx.Command(), Reset)
	}
}

// list prints all items in the file.
func list(file *os.File) {
	// Reset the file pointer to the beginning.
	_, err := file.Seek(0, io.SeekStart)
	check(err)
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	var todoLines []string
	for fileScanner.Scan() {
		todoLines = append(todoLines, fileScanner.Text())
	}
	if len(todoLines) == 0 {
		fmt.Printf("%sAll done!%s\n", Green, Reset)
	}
	for i, line := range todoLines {
		fmt.Printf("%s%d.%s %s\n", Green, i+1, Reset, line)
	}
}

// check panics if an error is not nil.
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

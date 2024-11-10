package main

import (
	"bufio"
	"fmt"
	"github.com/alecthomas/kong"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
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
		Title string `help:"Title of the todo item." arg:""`
	} `cmd:"" help:"Add a todo item."`

	Rm struct {
		Number []string `help:"Todo items to complete." arg:"" optional:""`
	} `cmd:"" help:"Complete one or more todo items. If no numbers are provided, complete all todo items."`
}

func main() {
	// Open the file for reading and writing.
	home, err := os.UserHomeDir()
	check(err)
	filename := filepath.Join(home, "dev", "todo", "todos.txt")
	_, err = os.Stat(filename)
	check(err)
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_RDWR, 0644)
	check(err)
	defer func(file *os.File) {
		err := file.Close()
		check(err)
	}(file)

	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "ls":
		list(file)

	case "add <title>":
		input := CLI.Add.Title
		_, err = file.Write([]byte(fmt.Sprintf("%s\n", input)))
		check(err)
		list(file)

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
			_, err := file.Write([]byte(fmt.Sprintf("%s\n", line)))
			check(err)
		}
		list(file)

	default:
		fmt.Println("Unknown command:", ctx.Command())
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

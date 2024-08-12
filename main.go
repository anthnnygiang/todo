package main

import (
	"bufio"
	"fmt"
	"github.com/alecthomas/kong"
	"io"
	"log"
	"os"
	"strconv"
)

var CLI struct {
	Add struct {
		Todo string `arg:"" help:"Name of todo item."`
	} `cmd:"" help:"Add a todo item."`

	List struct{} `cmd:"" help:"List all todo items."`

	Done struct {
		Number []string `help:"Todo items to complete" arg:"" optional:""`
	} `cmd:"" help:"Complete one or more todo items. If no numbers are provided, complete all todo items."`
}

// check panics if an error is not nil.
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Try to append to the file, if it doesn't exist, create it.
	file, err := os.OpenFile("todos.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	check(err)
	defer func(file *os.File) {
		err := file.Close()
		check(err)
	}(file)

	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "list":
		list(file)

	case "add <todo>":
		_, err = file.Write([]byte(fmt.Sprintf("%s\n", CLI.Add.Todo)))
		check(err)
		list(file)

	case "done":
		err := os.Truncate("todos.txt", 0)
		check(err)

	case "done <number>":
		numbersMap := make(map[int]bool)
		for _, n := range CLI.Done.Number {
			num, err := strconv.Atoi(n)
			check(err)
			numbersMap[num] = true
		}

		var newTodoLines []string
		fileScanner := bufio.NewScanner(file)
		fileScanner.Split(bufio.ScanLines)
		for i := 1; fileScanner.Scan(); i++ {
			if numbersMap[i] {
				continue
			}
			newTodoLines = append(newTodoLines, fileScanner.Text())
		}

		err = file.Truncate(0)
		check(err)

		for _, line := range newTodoLines {
			_, err := file.Write([]byte(fmt.Sprintf("%s\n", line)))
			check(err)
		}
		list(file)

	default:
		fmt.Println("Invalid command.")
	}
}

// list prints all items in the file.
func list(file *os.File) {
	_, err := file.Seek(0, io.SeekStart)
	check(err)
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	var todoLines []string
	for fileScanner.Scan() {
		todoLines = append(todoLines, fileScanner.Text())
	}
	for i, line := range todoLines {
		fmt.Printf("%d. %s\n", i+1, line)
	}
}

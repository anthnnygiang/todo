package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"strconv"
	"strings"
)

type Todo struct {
	title string
}

func main() {
	db, err := sql.Open("sqlite3", "./todos.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	lsCmd := flag.NewFlagSet("ls", flag.ExitOnError)
	doneCmd := flag.NewFlagSet("done", flag.ExitOnError)

	argLen := len(os.Args)
	if argLen < 2 {
		fmt.Println("Usage: todo [\"title\" | ls | done <i...>]")
		return
	}

	switch os.Args[1] {
	case "ls":
		// read all into memory
		// print out with array/slice i
		err := lsCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
		listTodos(db)

	case "done":
		// mark all with indexes as done
		// delete the ones with the indexes
		// write back to db in order

		// parse the indexes
		err := doneCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}

		// make a map of indexes
		var doneMap = make(map[int]bool)
		for _, arg := range doneCmd.Args() {
			index, err := strconv.Atoi(arg)
			if err != nil {
				log.Fatal(err)
			}
			doneMap[index] = true
		}

		// read todos from db
		var todos []Todo
		rows, err := db.Query("select title from todos")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		for rows.Next() {
			var todo Todo
			err = rows.Scan(&todo.title)
			if err != nil {
				log.Fatal(err)
			}
			todos = append(todos, todo)
		}
		err = rows.Err()
		if err != nil {
			log.Fatal(err)
		}

		// delete todos from db
		for i, todo := range todos {
			if doneMap[i+1] {
				// delete todos with the same title
				_, err := db.Exec("delete from todos where title = ?", todo.title)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		listTodos(db)

	case "clear":
		_, err := db.Exec("delete from todos")
		if err != nil {
			log.Fatal(err)
		}
		listTodos(db)

	default:
		_, err := db.Exec("insert into todos (title) values (?)", strings.Join(os.Args[1:], " "), false)
		if err != nil {
			log.Fatal(err)
		}
		listTodos(db)
	}
}

func listTodos(db *sql.DB) {
	rows, err := db.Query("select title from todos")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var index int = 1
	for rows.Next() {
		var todo Todo
		err = rows.Scan(&todo.title)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%d. %+v\n", index, todo.title)
		index++
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

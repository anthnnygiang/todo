package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

type Todo struct {
	title string
	done  bool
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
	if argLen < 2 || argLen > 2 {
		fmt.Println("Usage: todo [\"title\" | ls | done <...index>]")
		return
	}

	switch os.Args[1] {
	case "ls":
		// read all into memory
		// print out with array/slice index
		err := lsCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
		listTodos(db)

	case "done":
		// mark all with indexes as done
		// delete the ones with the indexes
		// write back to db in order

		err := doneCmd.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("mark as done", doneCmd.Args())
		listTodos(db)

	default:
		_, err := db.Exec("insert into todos (title, done) values (?, ?)", os.Args[1], false)
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

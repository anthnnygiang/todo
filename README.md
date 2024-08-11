# Todo CLI

## Setup
Create a SQLite database store
`$ sqlite3 todo.db < init.sql`

* `todo "todo"`: Add a todo
* `todo ls`: List todos
* `todo done ...id`: Complete a todo 

Example usage:
```
$ todo "do something"

$ todo ls
1. do something

$ todo done 1
```
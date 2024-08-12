# Todo CLI

## Installation
`$ go install`

## Usage
* `todo add "buy strawberries"`: Add a todo item
* `todo list`: Print all todo items
* `todo done [number ...]`: Complete todo items with the given numbers

Example usage:
```
$ todo add "buy strawberries"
$ todo add "buy bananas"

$ todo list
1. buy strawberries
2. buy bananas

$ todo done 1
1. buy bananas
```
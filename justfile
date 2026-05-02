bin := "todo"

default:
    just --list

test:
    go test ./...

build:
    go build -o {{bin}} .

install:
    go install .

clean:
    rm -f {{bin}}

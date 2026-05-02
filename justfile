bin := "todo"
coverage := "coverage.out"
coverage_html := "coverage.html"

default:
    just --list

test:
    go test ./...

coverage:
    go test -coverprofile={{coverage}} ./...
    go tool cover -func={{coverage}}

visualise: coverage
    go tool cover -html={{coverage}} -o {{coverage_html}}

build:
    go build -o {{bin}} .

install:
    go install .

clean:
    rm -f {{bin}}

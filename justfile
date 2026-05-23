bin := "todo"
coverage := "coverage.out"
coverage_html := "coverage.html"

# @just --list
default:
    @just --list

# go test ./...
test:
    go test ./...

# go fmt ./...
format:
    go fmt ./...

# go test -coverprofile={{ coverage }} ./...
coverage:
    go test -coverprofile={{ coverage }} ./...
    go tool cover -func={{ coverage }}

# go tool cover -html={{ coverage }} -o {{ coverage_html }}
visualise: coverage
    go tool cover -html={{ coverage }} -o {{ coverage_html }}

# go build -o {{ bin }} .
build:
    go build -o {{ bin }} .

# go install .
install:
    go install .

# rm -f {{ bin }}
clean:
    rm -f {{ bin }}

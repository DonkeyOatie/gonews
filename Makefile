all: deps build test lint

ci: test lint

deps:
	go get golang.org/x/tools/cmd/goimports
	go get github.com/stretchr/testify
	go get -u github.com/golang/lint/golint
	goimports -w .
	go get .

test:
	go test -coverprofile=coverage.out

cover: test
	go tool cover -func=coverage.out

coverhtml: test
	go tool cover -html=coverage.out

build: clean
	go build

clean:
	-rm gonews

lint:
	golint .

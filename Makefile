ifneq (,${KARMA_DOTENV})
	include ${KARMA_DOTENV}
	export
endif

COMMIT_SHA="$(shell git rev-parse --short HEAD)"

test:
	go test ./...
.PHONY: test

run:
	go run main.go
.PHONY: run

build:
	CGO_ENABLED=1 go build -o /karmabot .

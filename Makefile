ifneq (,${KARMA_DOTENV})
	include ${KARMA_DOTENV}
	export
endif

test:
	go test ./...
.PHONY: test

run:
	go run main.go
.PHONY: run

bin:
	go build -o karmabot_$(VERSION)

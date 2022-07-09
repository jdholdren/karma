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

# Needed to properly compile the binary from mac/other os'es not linux
compiler-image:
	docker build . -t karma-compiler

bin: compiler-image
	docker run -v $(shell pwd):/karma karma-compiler
.PHONY: bin

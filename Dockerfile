FROM golang:1.18

ARG VERSION

WORKDIR /karma/

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

ENTRYPOINT ["go", "build", "-o", "karmabot"]

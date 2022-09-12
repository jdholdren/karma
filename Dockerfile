FROM golang:1.18 AS builder

WORKDIR /karma

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o /karmabot .

CMD ["/karmabot"]

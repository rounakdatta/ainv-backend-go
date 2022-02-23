FROM golang:1.15

WORKDIR /app
COPY go.* ./
RUN go mod download

COPY . ./
RUN go build src/ainv/ainv.go

CMD ["/app/ainv", "ainv", "1234"]

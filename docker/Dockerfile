FROM golang:1.23 AS build

WORKDIR /app

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build runner
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /hoarderbot cmd/main.go

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /hoarderbot .

ENTRYPOINT ["/hoarderbot"]

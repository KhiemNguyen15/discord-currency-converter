FROM golang:1.22-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp .

FROM alpine:edge
LABEL org.opencontainers.image.source=https://github.com/khiemnguyen15/discord-currency-converter
WORKDIR /app
COPY --from=build /app/myapp .
COPY bcp47.csv .
ENTRYPOINT ["/app/myapp"]

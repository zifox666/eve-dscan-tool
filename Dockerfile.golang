FROM golang:1.24-alpine AS build

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN go build -o /out/eve-dscan-tool ./cmd/server

FROM alpine:3.22

WORKDIR /app

COPY --from=build /out/eve-dscan-tool /app/eve-dscan-tool

EXPOSE 8000
CMD ["/app/eve-dscan-tool"]


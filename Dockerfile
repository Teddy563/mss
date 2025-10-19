# syntax=docker/dockerfile:1

##
## Build
##

FROM golang:1.25-bookworm AS build

WORKDIR /mineplus

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY lib/ ./lib/

RUN go build -o /mineplus-proxy

##
## Deploy
##

FROM gcr.io/distroless/base-debian11

WORKDIR /app

COPY README.md ./README.md
COPY mineplus-config.json ./mineplus-config.json
COPY --from=build /mineplus-proxy ./mineplus-proxy

EXPOSE 25555

USER nonroot:nonroot

ENTRYPOINT ["/app/mineplus-proxy"]

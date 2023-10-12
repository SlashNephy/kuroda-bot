# syntax=docker/dockerfile:1
FROM golang:1.21-bullseye@sha256:b571a59dd275f99531870fad070ca37948a6017c4a050e492b61d786d62712e7 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ../.. ./
RUN make build

FROM debian:bullseye-slim@sha256:9bec46ecd98ce4bf8305840b021dda9b3e1f8494a0768c407e2b233180fa1466
WORKDIR /app

RUN groupadd -g 1000 app && useradd -u 1000 -g app app

RUN apt-get update \
    && apt-get install -yqq --no-install-recommends \
      ca-certificates \
    && rm -rf /var/lib/apt/lists/*

USER app
COPY --from=build /app/kuroda-bot ./
CMD ["./kuroda-bot"]

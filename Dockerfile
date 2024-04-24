# syntax=docker/dockerfile:1@sha256:dbbd5e059e8a07ff7ea6233b213b36aa516b4c53c645f1817a4dd18b83cbea56
FROM golang:1.22-bullseye@sha256:3b55bf3986b2651a515f48ddf758b80a7d78d8be18706fd98aa74241992dac96 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN make build

FROM debian:bullseye-slim@sha256:715354035496a48b9c4c8f146a6f751de70449913773038776eb1f3d01c93989
WORKDIR /app

RUN groupadd -g 1000 app && useradd -u 1000 -g app app

RUN apt-get update \
    && apt-get install -yqq --no-install-recommends \
      ca-certificates \
    && rm -rf /var/lib/apt/lists/*

USER app
COPY --from=build /app/kuroda-bot ./
CMD ["./kuroda-bot"]

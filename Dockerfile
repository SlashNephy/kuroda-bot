# syntax=docker/dockerfile:1@sha256:4c68376a702446fc3c79af22de146a148bc3367e73c25a5803d453b6b3f722fb
FROM golang:1.24-bullseye@sha256:ab9b6de58c573ffd99dbc2be54bbc9a206eb736f6be491fc07d3d2cde2c2cfe7 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN make build

FROM debian:bullseye-slim@sha256:e4b93db6aad977a95aa103917f3de8a2b16ead91cf255c3ccdb300c5d20f3015
WORKDIR /app

RUN groupadd -g 1000 app && useradd -u 1000 -g app app

RUN apt-get update \
    && apt-get install -yqq --no-install-recommends \
      ca-certificates \
    && rm -rf /var/lib/apt/lists/*

USER app
COPY --from=build /app/kuroda-bot ./
CMD ["./kuroda-bot"]

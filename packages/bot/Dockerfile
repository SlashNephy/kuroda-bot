# syntax=docker/dockerfile:1
FROM golang:1.21-bullseye@sha256:84af90fc39c7ff39e545f4859a2a5fc172f8669b89b535b43b621f9fa05dddb9 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ../.. ./
RUN make build

FROM debian:bullseye-slim@sha256:61386e11b5256efa33823cbfafd668dd651dbce810b24a8fb7b2e32fa7f65a85
WORKDIR /app

RUN groupadd -g 1000 app && useradd -u 1000 -g app app

RUN apt-get update \
    && apt-get install -yqq --no-install-recommends \
      ca-certificates \
    && rm -rf /var/lib/apt/lists/*

USER app
COPY --from=build /app/kuroda-bot ./
CMD ["./kuroda-bot"]

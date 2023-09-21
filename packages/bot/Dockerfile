# syntax=docker/dockerfile:1
FROM golang:1.21-bullseye@sha256:8fcb633c608b7f0b6415a7349a748ce3b8d357d1184d5e5a9e42cbcb77b78444 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ../.. ./
RUN make build

FROM debian:bullseye-slim@sha256:c618be84fc82aa8ba203abbb07218410b0f5b3c7cb6b4e7248fda7785d4f9946
WORKDIR /app

RUN groupadd -g 1000 app && useradd -u 1000 -g app app

RUN apt-get update \
    && apt-get install -yqq --no-install-recommends \
      ca-certificates \
    && rm -rf /var/lib/apt/lists/*

USER app
COPY --from=build /app/kuroda-bot ./
CMD ["./kuroda-bot"]

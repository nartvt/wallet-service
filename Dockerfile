FROM golang:1.21 AS builder

COPY . /src
WORKDIR /src

ARG GITHUB_ACCESS_USER
ARG GITHUB_ACCESS_TOKEN

RUN git config --global url."https://${GITHUB_ACCESS_USER}:${GITHUB_ACCESS_TOKEN}@github.com/".insteadOf "https://github.com/" && \
    go mod download && \
    git config --global --unset url."https://${GITHUB_ACCESS_USER}:${GITHUB_ACCESS_TOKEN}@github.com/".insteadOf


RUN make build

FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /src/bin /app
COPY configs /data/conf

WORKDIR /app

EXPOSE 8000
EXPOSE 9000

CMD ["./server", "-conf", "/data/conf/config.yaml"]

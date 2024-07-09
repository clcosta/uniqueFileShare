FROM golang:1.22.4 AS BUILDER
WORKDIR /app
COPY . /app/

RUN go build -o /app/shareFile ./src/

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=BUILDER /app/src /app/src
COPY --from=BUILDER /app/shareFile /app/shareFile

ENTRYPOINT ["/app/shareFile"]
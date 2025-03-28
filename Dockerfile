
FROM golang:1.24-bookworm AS builder
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64
RUN apt-get update && apt-get install -y gcc
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o server .
FROM debian:bookworm-slim
WORKDIR /root/
COPY --from=builder /app/server .
COPY ./public ./public  
EXPOSE 8080
RUN mkdir /data
VOLUME /data
CMD ["./server"]

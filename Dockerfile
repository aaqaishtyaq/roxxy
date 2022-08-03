FROM golang:1.19.0 as builder
WORKDIR /app
COPY . .
RUN go build

FROM alpine:latest
COPY --from=builder /app/roxxy /bin/roxxy
EXPOSE 8080
ENTRYPOINT [ "/bin/roxxy" ]

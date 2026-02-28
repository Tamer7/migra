FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY migra /usr/local/bin/migra

ENTRYPOINT ["/usr/local/bin/migra"]
CMD ["--help"]

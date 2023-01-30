FROM golang:1.19.5-alpine3.17 as builder
RUN apk update && apk add gcc musl-dev
WORKDIR /src
COPY . .
RUN go build  -o build/app cmd/healthcheck/main.go

FROM alpine:3.17
WORKDIR /bin
COPY --from=builder /src/build/app app
COPY --from=builder /src/config/config.json /var/config.json
RUN mkdir tmp
ENV APP.DB_CONN_STR=file:/var/storage.sqlite?cache=shared
ENTRYPOINT [ "app" ]
CMD ["-c", "/var/config.json"]

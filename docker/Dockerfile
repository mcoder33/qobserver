# Собираем в гошке
FROM golang:1.22 as build

ENV BIN_FILE /opt/qobserver/qobserver-app
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

# Собираем статический бинарник Go (без зависимостей на Си API),
# иначе он не будет работать в alpine образе.
ARG LDFLAGS
RUN CGO_ENABLED=0 go docker \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} cmd/*

# На выходе тонкий образ
FROM alpine:3.9

LABEL ORGANIZATION="CodingForFun"
LABEL SERVICE="qobserver"
LABEL MAINTAINERS="coder33@yandex.ru"

ENV BIN_FILE "/opt/qobserver/qobserver-app"
COPY --from=build ${BIN_FILE} ${BIN_FILE}

CMD ${BIN_FILE}

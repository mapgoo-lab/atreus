FROM  zyxtomcat/grpc-go:v1.0 as builder

WORKDIR /build
COPY . /build

RUN go env -w GOPROXY=https://goproxy.cn,direct \
    && export GOPROXY=https://goproxy.cn \
    && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./cmd/main.go \
    && rm -rf /etc/localtime \
    && ln -s  /usr/share/zoneinfo/Hongkong /etc/localtime \
    && dpkg-reconfigure -f noninteractive tzdata

FROM alpine:latest
MAINTAINER zhuyx <zhuyx@mapgoo.net>

RUN apk update  \
    && apk --no-cache add ca-certificates \
    && update-ca-certificates \
    && apk add --no-cache libc6-compat \
    && echo "hosts: files dns" > /etc/nsswitch.conf \
    && rm -rf /var/cache/apk/* \
    && mkdir /app

WORKDIR /app

COPY --from=builder /build/app .
COPY --from=builder /build/configs/ ./configs/

EXPOSE 8100 9100

CMD ["./app", "-conf", "./configs/"]

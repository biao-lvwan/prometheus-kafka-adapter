FROM golang:1.20.6-alpine as build
WORKDIR /src/prometheus-kafka-adapter

COPY go.mod go.sum *.go ./
ADD . /src/prometheus-kafka-adapter
#设置下载包
RUN sed -i -e 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache gcc musl-dev
# 设置代理服务器
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod tidy && go mod vendor
#-gcflags "all=-N -l" 为设置断点调试
RUN go build  -ldflags='-w -s -extldflags "-static"' -tags musl,static,netgo -mod=vendor -o /prometheus-kafka-adapter

FROM alpine:3.18

COPY schemas/metric.avsc /schemas/metric.avsc
COPY --from=build /prometheus-kafka-adapter /

CMD /prometheus-kafka-adapter

# 官方来源, 该映像的$GOPATH值已被设置为/go
FROM golang

# 安装etcd依赖
RUN go get github.com/coreos/etcd/clientv3
RUN go get gopkg.in/yaml.v2
RUN go get github.com/valyala/fasthttp
RUN mkdir -p /go/src/goagent

# 复制 当前目录下内容 到容器中
COPY . /go/src/goagent/

COPY --from=builder /usr/local/bin/docker-entrypoint.sh /usr/local/bin
COPY --from=builder /root/workspace/services/mesh-provider/target/mesh-provider-1.0-SNAPSHOT.jar /root/dists/mesh-provider.jar
COPY --from=builder /root/workspace/services/mesh-consumer/target/mesh-consumer-1.0-SNAPSHOT.jar /root/dists/mesh-consumer.jar

COPY start-agent.sh /usr/local/bin

RUN set -ex \
 && chmod a+x /usr/local/bin/start-agent.sh \
 && mkdir -p /root/logs

RUN go build -o /go/src/goagent/run /go/src/goagent/main.go

RUN chmod a+x /go/src/goagent/run

ENTRYPOINT ["start-agent.sh"]

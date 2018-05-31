# Builder container
FROM registry.cn-hangzhou.aliyuncs.com/aliware2018/services AS builder

# 官方来源, 该映像的$GOPATH值已被设置为/go
FROM golang

#  安装jdk8
RUN mkdir /var/tmp/jdk
RUN wget --no-check-certificate --no-cookies --header "Cookie: oraclelicense=accept-securebackup-cookie"  -P /var/tmp/jdk http://download.oracle.com/otn-pub/java/jdk/8u171-b11/512cd62ec5174c3487ac17c61aaa89e8/jdk-8u171-linux-x64.tar.gz
RUN tar xzf /var/tmp/jdk/jdk-8u171-linux-x64.tar.gz -C /var/tmp/jdk && rm -rf /var/tmp/jdk/jdk-8u171-linux-x64.tar.gz
#设置环境变量
ENV JAVA_HOME /var/tmp/jdk/jdk1.8.0_171
ENV PATH $PATH:$JAVA_HOME/bin


# 安装etcd依赖
RUN go get github.com/coreos/etcd/clientv3
RUN go get gopkg.in/yaml.v2
RUN go get github.com/valyala/fasthttp
#RUN go get github.com/ivpusic/neo
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

# golang.org/x/ 安装
#RUN git clone https://github.com/golang/net.git $GOPATH/src/github.com/golang/net
#RUN git clone https://github.com/golang/text.git $GOPATH/src/github.com/golang/text
#RUN mkdir -p $GOPATH/src/golang.org/
#RUN ln -sf $GOPATH/src/github.com/golang $GOPATH/src/golang.org/x
#RUN go install text
#RUN go get golang.org/x/net/http2


RUN go build -o /go/src/goagent/main /go/src/goagent/main.go

RUN chmod a+x /go/src/goagent/main


# Expose the application on port 8087
EXPOSE 8087

ENTRYPOINT ["docker-entrypoint.sh"]

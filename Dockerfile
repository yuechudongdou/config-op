FROM 172.16.2.18:5000/golang:1.17 as build

# 容器环境变量添加，会覆盖默认的变量值
ENV GO111MODULE=auto
ENV GOPROXY=https://goproxy.cn,direct

# 设置工作区
WORKDIR /go/release

# 把全部文件添加到/go/release目录
ADD . .

# 编译：把cmd/main.go编译成可执行的二进制文件，命名为app
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -o configOp main.go

# 运行：使用scratch作为基础镜像
FROM 172.16.2.18:5000/alpine as prod

# 在build阶段复制时区到
WORKDIR /go/release/
COPY --from=build /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
# 在build阶段复制可执行的go二进制文件app
COPY --from=build /go/release/configOp /usr/bin/configOp
# 在build阶段复制配置文件
# 启动服务
ENV GIN_MODE=release
ENTRYPOINT ["configOp"]
CMD ["help"]

# 使用Go 1.20镜像作为基础镜像
FROM golang:1.20
ENV GROOT=/go-nginx
# 使用RUN指令安装Nginx
RUN apt-get update && apt-get install -y nginx net-tools curl
RUN mkdir -p ${GROOT} && mkdir -p /var/log/go
# 将工作目录设置为Nginx的根目录
WORKDIR ${GROOT}
ENV NGINX_XML=/etc/nginx
# 将Go应用程序的代码复制到容器中
COPY ./bin .
ENTRYPOINT [ "entrypoint.sh" ]
# 使用CMD指令启动Nginx
CMD [ "-d","on" ]

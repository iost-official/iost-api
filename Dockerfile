FROM iostio/iost-dev as builder
ENV GOPATH /gopath
WORKDIR $GOPATH/src/github.com/iost-official/iost-api
COPY . ./
RUN make && cd task && make

FROM alpine as base
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

FROM base as iost-api
ENV GOPATH /gopath
WORKDIR $GOPATH/src/github.com/iost-official/iost-api
COPY --from=builder $GOPATH/src/github.com/iost-official/iost-api/iost-api .
EXPOSE 8002
CMD ["./iost-api"]

FROM base as iost-api-task
ENV GOPATH /gopath
WORKDIR $GOPATH/src/github.com/iost-official/iost-api/task
COPY --from=builder $GOPATH/src/github.com/iost-official/iost-api/task/iost-api-task .
CMD ["./iost-api-task"]

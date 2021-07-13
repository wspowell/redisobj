FROM golang:1.16

RUN apt-get update
RUN apt-get -y install redis-tools
RUN rm -rf /var/lib/apt/lists/*

RUN GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.41.1
RUN GO111MODULE=on go get github.com/uudashr/gopkgs/v2/cmd/gopkgs@v2.1.2
# The version is correct for go-outline. The tag is not prefixed with 'v'.
RUN GO111MODULE=on go get github.com/ramya-rao-a/go-outline@1.0.0
RUN GO111MODULE=on go get github.com/cweill/gotests/...@v1.6.0
RUN GO111MODULE=on go get github.com/fatih/gomodifytags@v1.13.0
RUN GO111MODULE=on go get github.com/josharian/impl@v1.1.0
RUN GO111MODULE=on go get github.com/haya14busa/goplay/cmd/goplay@v1.0.0
RUN GO111MODULE=on go get github.com/go-delve/delve/cmd/dlv@v1.6.1
RUN GO111MODULE=on go get golang.org/x/tools/gopls@v0.7.0
RUN GO111MODULE=on go get github.com/vektra/mockery/v2/.../@v2.9.0
RUN GO111MODULE=on go get github.com/rakyll/statik@v0.1.7
RUN GO111MODULE=on go get gotest.tools/gotestsum@v1.6.4
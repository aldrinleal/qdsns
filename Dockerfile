FROM golang:1.23-bullseye as builder

RUN apt-get update && apt-get -y install upx-ucl build-essential && rm -rf /var/lib/apt/lists/*

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/github.com/aldrinleal/qdsns

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN mkdir -p /app/bin && \
    CGO_ENABLED=0 GOOS=linux go build -v -a -ldflags='-w -s' -o $GOPATH/bin/qdsns ./cmd/qdsns && \
    upx $GOPATH/bin/* && \
    cp -v $GOPATH/bin/* /app/bin/

FROM alpine
# This container exposes port 8080 to the outside world

COPY --from=builder /app /app

WORKDIR /app

ENV PORT=5000

EXPOSE 5000

# Run the executable
CMD ["/app/bin/qdsns"]

FROM golang:alpine AS builder
RUN apk add git
WORKDIR $GOPATH/src/app
RUN go get -u github.com/gorilla/mux
RUN go get -u gitlab.com/golang-commonmark/markdown
ADD . .
RUN go run main.go clean build
RUN go get -u github.com/m3ng9i/ran
WORKDIR $GOPATH/src/github.com/m3ng9i/ran
RUN go build -o ran
RUN echo "Tip: Copy binary from here: $GOPATH/src/app/build" 

FROM alpine
WORKDIR /app
COPY --from=builder /go/src/app/build/ ./
COPY --from=builder /go/src/github.com/m3ng9i/ran . 
ENTRYPOINT ./ran

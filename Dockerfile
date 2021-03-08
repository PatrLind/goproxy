FROM golang:1.16-alpine as builder
RUN mkdir /build
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN go build -o goproxy

FROM golang:1.16-alpine
RUN apk update && apk add ca-certificates git
RUN rm -rf /var/cache/apk/*
RUN mkdir /cache

WORKDIR /goproxy
COPY --from=builder /build/goproxy ./

CMD ["./goproxy"]

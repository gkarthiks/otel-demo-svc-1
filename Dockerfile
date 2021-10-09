FROM golang:1.17-alpine AS builder
WORKDIR /go/src/app
COPY . /go/src/app
RUN go build main.go


FROM alpine
COPY --from=builder /go/src/app/main /bin
CMD ["/bin/main"]
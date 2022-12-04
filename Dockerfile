FROM golang:1.19.3 as builder
WORKDIR /go/src/app
COPY . .
RUN go build -o /go/bin/app ./...

#use distroless image
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /go/bin/app /
RUN  app
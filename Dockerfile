FROM golang:alpine AS build

WORKDIR /build

ADD . /build

RUN apk --no-cache add git ca-certificates && update-ca-certificates

RUN go get ./...

ENV GOOS=linux
ENV GOARCH=amd64

RUN go build -o globalflow

FROM alpine

RUN apk --no-cache add gpg
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/globalflow /

ENTRYPOINT ["/globalflow"]
CMD ["server"]
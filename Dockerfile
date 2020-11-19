FROM golang:1.14-alpine3.12 as build
RUN apk update && apk add --no-cache git
WORKDIR /go/src/app

COPY /static ./static
COPY /templates ./templates
COPY /*.go ./

RUN go get -d -v
RUN go build -o /app

FROM alpine:3.12
WORKDIR /usr/bin
COPY /static ./static
COPY /templates ./templates
COPY --from=build /app .
EXPOSE 8080
CMD ["app"]
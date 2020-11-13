FROM golang:1.14-alpine
RUN apk update && apk add --no-cache git
WORKDIR /go/src/app

COPY /static ./static
COPY /templates ./templates
COPY /*.go ./

RUN go get -d -v
RUN go install -v

EXPOSE 8080

CMD ["app"]
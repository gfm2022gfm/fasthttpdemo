FROM golang:alpine AS build-env
ADD . /src
#RUN go env -w GOPROXY=https://goproxy.cn
RUN cd /src && ls && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app

#RUN apk add curl
#ENTRYPOINT ["/app"]



#FROM centurylink/ca-certs
FROM alpine
RUN apk add curl
COPY --from=build-env /src/app /
ENTRYPOINT ["/app"]


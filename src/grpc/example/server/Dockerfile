FROM golang:alpine AS build-env
ADD . /src
#RUN cd /src/server && ls
RUN cd /src/server && ls && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app
RUN ls


#FROM centurylink/ca-certs
FROM alpine
RUN apk add curl
COPY --from=build-env /src/server/app /
ENTRYPOINT ["/app"]


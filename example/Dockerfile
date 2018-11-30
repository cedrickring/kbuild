FROM golang:alpine as build-env
COPY . /src/
RUN cd /src && go build -o helloworld main.go

FROM alpine
WORKDIR /app
COPY --from=build-env /src/helloworld /app/
CMD ["/app/helloworld"]
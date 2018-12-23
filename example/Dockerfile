FROM golang:alpine as build-env
ARG MAIN=main.go
COPY . /src/
RUN cd /src && go build -o helloworld $MAIN

FROM scratch
WORKDIR /app
COPY --from=build-env /src/helloworld /app/
ENTRYPOINT ["/app/helloworld"]
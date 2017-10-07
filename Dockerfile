FROM golang:alpine3.6 AS binary
ADD . /app
WORKDIR /app
RUN CGO_ENABLED=0 go build -o msgsrv

FROM scratch
MAINTAINER Andreas Krey <a.krey@gmx.de>

WORKDIR /data

COPY --from=binary /app/msgsrv /app

EXPOSE 3046

VOLUME ["/data"]
CMD ["/app/msgsrv"]

# docker run --name msgsrv -v /tmp/msgsrv-data:/data -v /etc/localtime:/etc/localtime:ro -p 3042:3046 -u `id -u` apky/msgsrv

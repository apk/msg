FROM scratch
MAINTAINER Andreas Krey <a.krey@gmx.de>

WORKDIR /data

COPY msgsrv /usr/bin/msgsrv

VOLUME ["/data"]
CMD ["/usr/bin/msgsrv"]

# docker run --name msgsrv -v /tmp/msgsrv-data:/data -v /etc/localtime:/etc/localtime:ro -p 3042:3046 -u `id -u` apky/msgsrv

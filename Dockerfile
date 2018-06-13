FROM alpine:latest
LABEL authors="Keith Hudgins <keith@docker.com>, Matt Smith <matthew.smith491@gmail.com>"

ARG home
ARG node1
ARG node2
RUN mkdir -p $home
COPY .docker/machine/machines/$node1 $home/.docker/machine/machines/$node1
COPY .docker/machine/machines/$node2 $home/.docker/machine/machines/$node2
COPY node1 node1
COPY node2 node2

COPY bin/vol-test /bin/usr/sbin/vol-test

CMD ["/bin/usr/sbin/vol-test"]

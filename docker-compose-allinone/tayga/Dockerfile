FROM alpine:edge

MAINTAINER Daneyon Hansen <danehans@cisco.com>

ENV \
	TAYGA_CONF_DATA_DIR=/var/db/tayga \
	TAYGA_CONF_DIR=/usr/local/etc \
	TAYGA_CONF_IPV4_ADDR=172.18.0.100 \
	TAYGA_CONF_PREFIX=2001:db8:64:ff9b::/96 \
	TAYGA_CONF_DYNAMIC_POOL=172.18.0.128/25

RUN apk add --no-cache tayga --repository=http://dl-cdn.alpinelinux.org/alpine/edge/testing/
RUN apk add --no-cache iptables

ADD docker-entry.sh /
RUN chmod +x /docker-entry.sh

ENTRYPOINT ["/docker-entry.sh"]
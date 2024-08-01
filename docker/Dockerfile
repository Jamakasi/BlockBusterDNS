FROM golang:alpine AS builder

LABEL stage=gobuilder
RUN apk update --no-cache && apk add --no-cache tzdata git ca-certificates make
WORKDIR /build
RUN git clone https://github.com/coredns/coredns.git \
	&& git clone https://github.com/Jamakasi/BlockBusterDNS.git \
	&& mv -f BlockBusterDNS/plugin/* coredns/plugin/
RUN sed '/dns64:dns64/a dns64_hack:dns64_hack' coredns/plugin.cfg >> coredns/plugin.cfg1 && mv coredns/plugin.cfg1 coredns/plugin.cfg \
	&& sed '/dns64_hack:dns64_hack/a cut_cname:cut_cname' coredns/plugin.cfg >> coredns/plugin.cfg1 && mv coredns/plugin.cfg1 coredns/plugin.cfg \
	&& sed '/local:local/a dmatcher:dmatcher' coredns/plugin.cfg >> coredns/plugin.cfg1 && mv coredns/plugin.cfg1 coredns/plugin.cfg \
	&& sed '/view:view/a finalize:github.com/tmeckel/coredns-finalizer' coredns/plugin.cfg >> coredns/plugin.cfg1 && mv coredns/plugin.cfg1 coredns/plugin.cfg
RUN make -C coredns/ \
	&& mkdir ../release \
	&& cp coredns/coredns ../release && chmod +x ../release/coredns \
	&& cp -r BlockBusterDNS/templates ../release \
	&& cp -r BlockBusterDNS/www_static ../release


FROM alpine
RUN apk update --no-cache && apk add --no-cache ca-certificates
COPY --from=builder /usr/share/zoneinfo/Europe/Moscow /usr/share/zoneinfo/Europe/Moscow
COPY --from=builder /release/ /dns/
ADD Corefile /dns/
ENV TZ Europe/Moscow
WORKDIR /dns

CMD ["./coredns"]
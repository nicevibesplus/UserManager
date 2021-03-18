ARG GOVERSION="1.16"
FROM golang:${GOVERSION}-alpine AS buildenv
ARG CGO_ENABLED="0"
ARG GOOS="linux"
ARG VERSION="0.1.0"

RUN apk add openssl
WORKDIR /build

COPY go.mod go.sum *.go vendor public ./
COPY vendor vendor
COPY public public

RUN go build -a -v -ldflags '-extldflags "-static"' -o usermanager .

# generate jwt keys
RUN mkdir keys && \
	openssl genrsa -out keys/jwt.key 4096 && \
	openssl rsa -in keys/jwt.key -pubout -out keys/jwt.pub

# # generate selfsigned tls cert
# RUN openssl req -x509 -sha256 -nodes -newkey rsa:2048 \
# 	-days 365 -subj "/C=DE/ST=Foobar/L=Baz/O=geofs/CN=localhost" \
# 	-keyout keys/tls.key -out keys/tls.crt

FROM scratch
ARG VERSION=${VERSION}
LABEL org.opencontainers.image.title="usermanager - geofs http frontend for LDAP"
LABEL org.opencontainers.image.description=""
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.authors="Norwin Roosen <bugs@nroo.de>"
LABEL org.opencontainers.image.vendor="geofs"
COPY --from=buildenv /build/keys /keys
COPY --from=buildenv /build/usermanager /

# required conf
# ENV UM_LDAP_ADMIN=
# ENV UM_LDAP_PASS=
# ENV UM_LDAP_BASE_DN=
# ENV UM_LDAP_ADMINFILTER=

# optional conf
# ENV UM_LDAP_SERVER=
# ENV UM_LDAP_PORT=
# ENV UM_LDAP_USERFILTER=
# ENV UM_JWT_PUB=
# ENV UM_JWT_PRIV=
# ENV UM_TLS_CERT=
# ENV UM_TLS_KEY=

EXPOSE 8443
CMD ["/usermanager"]

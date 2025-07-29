# SPDX-License-Identifier: Apache-2.0

########################################################################
##    docker build --no-cache --target certs -t vela-slack:certs .    ##
########################################################################

FROM alpine:3.22.1@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1 as certs

RUN apk add --update --no-cache ca-certificates

#########################################################
##    docker build --no-cache -t vela-slack:local .    ##
#########################################################

FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY release/vela-slack /bin/vela-slack

ENTRYPOINT [ "/bin/vela-slack" ]

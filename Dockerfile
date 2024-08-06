# SPDX-License-Identifier: Apache-2.0

########################################################################
##    docker build --no-cache --target certs -t vela-slack:certs .    ##
########################################################################

FROM alpine@sha256:0a4eaa0eecf5f8c050e5bba433f58c052be7587ee8af3e8b3910ef9ab5fbe9f5 as certs

RUN apk add --update --no-cache ca-certificates

#########################################################
##    docker build --no-cache -t vela-slack:local .    ##
#########################################################

FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY release/vela-slack /bin/vela-slack

ENTRYPOINT [ "/bin/vela-slack" ]

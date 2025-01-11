# Multi-stage docker that builds both the client and the server,
# runs tests
# and eventualy creates server image.
# (it's rather uncommon to create client images, although this is technically possible too)

FROM golang AS builder

RUN apt-get update && \
    apt-get install -y protobuf-compiler && \
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    go install go.uber.org/mock/mockgen@latest


WORKDIR /home/teleport
COPY proto/ proto/
COPY src/ src/
COPY certs/ certs/

RUN cd src && bash test.sh

RUN cd src/client && go generate && go build

RUN cd src/server && go generate && go build

FROM alpine AS server
RUN apk add libc6-compat
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
COPY --from=builder --chown=appuser /home/teleport/src/server/server /home/appuser/server
COPY --chown=appuser docker/entrypoint.sh /home/appuser/

ENTRYPOINT ["sh", "/home/appuser/entrypoint.sh"]
EXPOSE 1234



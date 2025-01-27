# ---------------------------------------------------------#
#                     Build web image                      #
# ---------------------------------------------------------#
FROM --platform=$BUILDPLATFORM node:16 as web

WORKDIR /usr/src/app

COPY web/package.json ./
COPY web/yarn.lock ./

# If you are building your code for production
# RUN npm ci --omit=dev

COPY ./web .

RUN yarn && yarn build && yarn cache clean

# ---------------------------------------------------------#
#                   Build Nxenv image                    #
# ---------------------------------------------------------#
FROM --platform=$BUILDPLATFORM golang:1.22-alpine3.18 as builder

RUN apk update \
    && apk add --no-cache protoc build-base git

# Setup workig dir
WORKDIR /app
RUN git config --global --add safe.directory '/app'

# Get dependencies - will also be cached if we won't change mod/sum
COPY go.mod .
COPY go.sum .

COPY Makefile .
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
RUN make dep
RUN make tools
# COPY the source code as the last step
COPY . .

COPY --from=web /usr/src/app/dist /app/web/dist

# build
ARG GIT_COMMIT
ARG RAPIDSIP_VERSION_MAJOR
ARG RAPIDSIP_VERSION_MINOR
ARG RAPIDSIP_VERSION_PATCH
ARG TARGETOS TARGETARCH

RUN if [ "$TARGETARCH" = "arm64" ]; then \
    wget -P ~ https://musl.cc/aarch64-linux-musl-cross.tgz && \
    tar -xvf ~/aarch64-linux-musl-cross.tgz -C ~ ; \
    fi

# set required build flags
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    if [ "$TARGETARCH" = "arm64" ]; then CC=~/aarch64-linux-musl-cross/bin/aarch64-linux-musl-gcc; fi && \
    LDFLAGS="-X github.com/nxenv/rapidship/version.GitCommit=${GIT_COMMIT} -X github.com/nxenv/rapidship/version.major=${RAPIDSIP_VERSION_MAJOR} -X github.com/nxenv/rapidship/version.minor=${RAPIDSIP_VERSION_MINOR} -X github.com/nxenv/rapidship/version.patch=${RAPIDSIP_VERSION_PATCH} -extldflags '-static'" && \
    CGO_ENABLED=1 \
    GOOS=$TARGETOS GOARCH=$TARGETARCH \
    CC=$CC go build -ldflags="$LDFLAGS" -o ./rapidship ./cmd/rapidship

### Pull CA Certs
FROM --platform=$BUILDPLATFORM alpine:latest as cert-image

RUN apk --update add ca-certificates

# ---------------------------------------------------------#
#                   Create final image                     #
# ---------------------------------------------------------#
FROM --platform=$TARGETPLATFORM alpine/git:2.43.0 as final

# setup app dir and its content
WORKDIR /app
VOLUME /data

ENV XDG_CACHE_HOME /data
ENV RAPIDSIP_GIT_ROOT /data
ENV RAPIDSIP_DATABASE_DRIVER sqlite3
ENV RAPIDSIP_DATABASE_DATASOURCE /data/database.sqlite
ENV RAPIDSIP_METRIC_ENABLED=true
ENV RAPIDSIP_METRIC_ENDPOINT=https://stats.drone.ci/api/v1/rapidship
ENV RAPIDSIP_TOKEN_COOKIE_NAME=token
ENV RAPIDSIP_DOCKER_API_VERSION 1.41
ENV RAPIDSIP_SSH_ENABLE=true
ENV RAPIDSIP_GITSPACE_ENABLE=true

COPY --from=builder /app/rapidship /app/rapidship
COPY --from=cert-image /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 3000
EXPOSE 3022

ENTRYPOINT [ "/app/rapidship", "server" ]

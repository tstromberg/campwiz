# Stage 1: Compile
FROM golang AS builder
WORKDIR /app
ENV SRC_DIR=/src/campwiz
ENV GO111MODULE=on
RUN mkdir -p ${SRC_DIR}/cmd ${SRC_DIR}/third_party ${SRC_DIR}/pkg ${SRC_DIR}/site /app/third_party /app/site
COPY go.* $SRC_DIR/
COPY cmd ${SRC_DIR}/cmd/
COPY pkg ${SRC_DIR}/pkg/
WORKDIR $SRC_DIR
RUN go mod download
RUN go build cmd/server/server.go

# Stage 2: Deploy
FROM gcr.io/distroless/base AS campwiz
COPY --from=builder /src/campwiz/server /app/
COPY site /app/site/
COPY metadata /app/metadata
#COPY third_party /app/third_party/

# Useful environment variables:
# 
# * PORT: Sets HTTP listening port (defaults to 8080)
# * PERSIST_BACKEND: Set the cache persistence backend
# * PERSIST_PATH: Set the cache persistence path
# 
# For other environment variables, see:
# https://github.com/google/campwiz/blob/master/docs/deploy.md
CMD ["/app/server", "--site=/app/site", "--3p=/app/third_party"]
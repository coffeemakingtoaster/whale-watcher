FROM golang:1.24-bookworm AS build

WORKDIR /build

COPY . .

# Install deps
RUN apt update && apt install -y python3-pip && \
python3 -m pip install pybindgen --break-system-packages && \
  go install golang.org/x/tools/cmd/goimports@latest && \
  go install github.com/go-python/gopy@latest

# Clean is not necessary here...but better safe than sorry
RUN make clean all verify

FROM debian AS runtime

WORKDIR /app

COPY --from=build /build/build/whale-watcher ./whale

ENTRYPOINT ["/app/whale"]






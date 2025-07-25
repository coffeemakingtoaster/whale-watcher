FROM golang:1.24-bookworm AS build

WORKDIR /build

# Install deps
RUN apt update && apt install -y python3-pip && \
python3 -m pip install pybindgen --break-system-packages && \
  go install golang.org/x/tools/cmd/goimports@latest && \
  go install github.com/go-python/gopy@latest

COPY . .

# Clean is not necessary here...but better safe than sorry
RUN make clean all 

FROM python:3.10-bookworm AS runtime

WORKDIR /app

COPY --from=build /build/build/whale-watcher ./whale-watcher

ENTRYPOINT ["/app/whale-watcher"]

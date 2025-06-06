cmd_lib: ./pkg/runner/command_util/library.so
fs_lib: ./pkg/runner/fs_util/library.so
os_lib: ./pkg/runner/os_util/library.so

./pkg/runner/command_util/library.so: ./pkg/runner/command_util/command_util.go.so
	go build -buildmode=c-shared -o ./pkg/runner/command_util/library.so ./pkg/runner/command_util/command_util.go.so

./pkg/runner/fs_util/library.so: ./pkg/runner/fs_util/fs_util.go.so
	go build -buildmode=c-shared -o ./pkg/runner/fs_util/library.so ./pkg/runner/fs_util/fs_util.go.so

./pkg/runner/os_util/library.so: ./pkg/runner/os_util/os_util.go.so
	go build -buildmode=c-shared -o ./pkg/runner/os_util/library.so ./pkg/runner/os_util/os_util.go.so

exec: ./whale-watcher

./whale-watcher: main.go
	go build -o ./whale-watcher ./cmd/whale-watcher/whale-watcher.go

all: cmd_lib fs_lib os_lib exec

clean:
	rm ./whale-watcher && rm ./pkg/runner/*_util/*.{so,h}

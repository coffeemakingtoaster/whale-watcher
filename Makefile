# Logging added by chatgpt
# Every other makefile war crime commited in here is my fault :)

# Define common paths
BUILD_DIR = ./build
CMD_DIR = ./cmd/whale-watcher
PKG_DIR = ./pkg/runner

# ANSI color codes
RED = \033[31m
GREEN = \033[32m
YELLOW = \033[33m
BLUE = \033[34m
PURPLE = \033[35m
RESET = \033[0m

DELIM = ************

# Phony targets declaration
.PHONY: all clean cmd_lib fs_lib os_lib exec docker

# Ensure the build directory exists
$(BUILD_DIR):
	@echo "\n$(BLUE)$(DELIM) Creating build directory $(DELIM)$(RESET)"
	mkdir -p $(BUILD_DIR)

cmd_lib: $(PKG_DIR)/command_util_build/__init__.py

# Targets for libraries with the build directory as a prerequisite
$(PKG_DIR)/command_util_build/__init__.py: $(PKG_DIR)/command_util/command_util.go
	@echo "\n$(PURPLE)$(DELIM) Building command_util library $(DELIM)$(RESET)"
	gopy build -output=$(PKG_DIR)/command_util_build -vm="python3" -rename=true $(PKG_DIR)/command_util/

fs_lib:  $(PKG_DIR)/fs_util_build/__init__.py

$(PKG_DIR)/fs_util_build/__init__.py: $(PKG_DIR)/fs_util/fs_util.go
	@echo "\n$(PURPLE)$(DELIM) Building fs_util library $(DELIM)$(RESET)"
	gopy build -output=$(PKG_DIR)/fs_util_build -vm="python3" -rename=true $(PKG_DIR)/fs_util

os_lib:  $(PKG_DIR)/os_util_build/__init__.py

$(PKG_DIR)/os_util_build/__init__.py: $(PKG_DIR)/os_util/os_util.go
	@echo "\n$(PURPLE)$(DELIM) Building os_util library $(DELIM)$(RESET)"
	gopy build -output=$(PKG_DIR)/os_util_build -vm="python3" -rename=true $(PKG_DIR)/os_util

# Target for the executable with the build directory as a prerequisite
exec: $(BUILD_DIR)/whale-watcher

$(BUILD_DIR)/whale-watcher: $(CMD_DIR)/whale-watcher.go | $(BUILD_DIR)
	@echo "\n$(PURPLE)$(DELIM) Building whale-watcher executable $(DELIM)$(RESET)"
	go build -o $(BUILD_DIR)/whale-watcher $(CMD_DIR)/whale-watcher.go

# Define the main target
all: cmd_lib fs_lib os_lib exec
	@echo "\n$(GREEN)$(DELIM) All targets built successfully! $(DELIM)$(RESET)"

# Clean target to remove the build directory
clean:
	@echo "\n$(RED)$(DELIM) Cleaning build directory $(DELIM)$(RESET)"
	rm -rf $(BUILD_DIR)
	rm -rf $(PKG_DIR)/*_build
	rm -rf ./out

docker:
	docker build -t whale-watcher:latest .

test:
	go test ./...

./out/out.tar:
	mkdir -p out
	docker buildx create --driver docker-container --driver-opt image=moby/buildkit:master,network=host --use
	docker buildx build -o type=oci,dest=./out/out.tar,compression=gzip .

oci-export: ./out/out.tar

# Run test ruleset that doesn't need a container but performs a basic signature check for the utils
verify: all test oci-export
	@echo "\n$(BLUE)$(DELIM) Verifying ruleset $(DELIM)$(RESET)"
	./build/whale-watcher $$(pwd)/_example/verify_ruleset.yaml $$(pwd)/Dockerfile "./out/out.tar"

docker-verify: docker
	docker run --rm -v $$(pwd)/_example/verify_ruleset.yaml:/app/verify_ruleset.yaml -v $$(pwd)/Dockerfile:/app/Dockerfile -it whale-watcher:latest "/app/verify_ruleset.yaml" "/app/Dockerfile" "whale-watcher:latest"

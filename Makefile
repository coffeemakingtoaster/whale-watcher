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

# Get the version from the latest git tag or default to "alpha"
VERSION = $(shell git describe --tags --abbrev=0 2>/dev/null || echo "alpha")

# Phony targets declaration
.PHONY: help
help:
	@echo "Whale Watcher - Version: $(VERSION)"
	@echo "Available targets:"
	@echo "  dep-install   - Install necessary dependencies."
	@echo "  clean         - Remove build artifacts."
	@echo "  all           - Build all targets."
	@echo "  verify        - Verify ruleset."
	@echo "\t--- internal ---"
	@echo "  cmd_lib       - Build the command_util library."
	@echo "  fs_lib        - Build the fs_util library."
	@echo "  os_lib        - Build the os_util library."
	@echo "  fix_lib        - Build the os_util library."
	@echo "  exec          - Build the whale-watcher executable."
	@echo "  test          - Run tests."
	@echo "  oci-export    - Export OCI image."
	@echo "  verify-local  - Verify ruleset with local ruleset"
	@echo "  verify-remote  - Verify ruleset loaded from git"
	@echo "\t--- not supported yet ---"
	@echo "  docker        - Build the Docker image."
	@echo "  docker-verify - Verify ruleset using Docker."

.PHONY: dep-install
dep-install:
	@echo "$(RED)Pip install uses break system packages...this should not be problem. If it is, don't blame me :)$(RESET)"
	python3 -m pip install pybindgen --break-system-packages
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/go-python/gopy@04cb87bab1e03ba3d50e827654e6e07c94ed7606

# Ensure the build directory exists
$(BUILD_DIR):
	@echo "\n$(BLUE)$(DELIM) Creating build directory $(DELIM)$(RESET)"
	mkdir -p $(BUILD_DIR)

.PHONY: cmd_lib
cmd_lib: $(PKG_DIR)/_command_util_build/__init__.py

# Targets for libraries with the build directory as a prerequisite
$(PKG_DIR)/_command_util_build/__init__.py: $(PKG_DIR)/command_util/command_util.go
	@echo "\n$(PURPLE)$(DELIM) Building command_util library $(DELIM)$(RESET)"
	gopy build -output=$(PKG_DIR)/_command_util_build -vm="python3" -rename=true $(PKG_DIR)/command_util/

.PHONY: fs_lib
fs_lib: $(PKG_DIR)/_fs_util_build/__init__.py

$(PKG_DIR)/_fs_util_build/__init__.py: $(PKG_DIR)/fs_util/fs_util.go
	@echo "\n$(PURPLE)$(DELIM) Building fs_util library $(DELIM)$(RESET)"
	gopy build -output=$(PKG_DIR)/_fs_util_build -vm="python3" -rename=true $(PKG_DIR)/fs_util

.PHONY: os_lib
os_lib: $(PKG_DIR)/_os_util_build/__init__.py

$(PKG_DIR)/_os_util_build/__init__.py: $(PKG_DIR)/os_util/os_util.go
	@echo "\n$(PURPLE)$(DELIM) Building os_util library $(DELIM)$(RESET)"
	gopy build -output=$(PKG_DIR)/_os_util_build -vm="python3" -rename=true $(PKG_DIR)/os_util

.PHONY: fix_lib
os_lib: $(PKG_DIR)/_fix_util_build/__init__.py

$(PKG_DIR)/_fix_util_build/__init__.py: $(PKG_DIR)/fix_util/fix_util.go
	@echo "\n$(PURPLE)$(DELIM) Building fix_util library $(DELIM)$(RESET)"
	gopy build -output=$(PKG_DIR)/_fix_util_build -vm="python3" -rename=true $(PKG_DIR)/fix_util

.PHONY: exec
exec: $(BUILD_DIR)/whale-watcher

$(BUILD_DIR)/whale-watcher: $(CMD_DIR)/whale-watcher.go | $(BUILD_DIR)
	@echo "\n$(PURPLE)$(DELIM) Building whale-watcher executable $(DELIM)$(RESET)"
	go build -o $(BUILD_DIR)/whale-watcher $(CMD_DIR)/whale-watcher.go

.PHONY: all
# Define the main target
all: cmd_lib fs_lib os_lib exec
	@echo "\n$(GREEN)$(DELIM) All targets built successfully! $(DELIM)$(RESET)"

.PHONY: clean
# Clean target to remove the build directory
clean:
	@echo "\n$(RED)$(DELIM) Cleaning build directory $(DELIM)$(RESET)"
	rm -rf $(BUILD_DIR)
	rm -rf $(PKG_DIR)/*_build
	rm -rf ./out

.PHONY: docker
docker:
	docker build -t whale-watcher:latest .

.PHONY: test
test:
	go test ./...

./out/out.tar:
	mkdir -p out
	docker buildx create --driver docker-container --driver-opt image=moby/buildkit:master,network=host --use
	docker buildx build -o type=oci,dest=./out/out.tar,compression=gzip -f ./_example/example.Dockerfile ./_example/
	docker buildx prune -a -f

.PHONY: oci-export
oci-export: ./out/out.tar

./out/out_docker.tar:
	mkdir -p out
	docker buildx create --driver docker-container --driver-opt image=moby/buildkit:master,network=host --use
	docker buildx build -o type=tar,dest=./out/out_docker.tar,compression=gzip -f ./_example/example.Dockerfile ./_example/
	docker buildx prune -a -f

.PHONY: docker-export
docker-export: ./out/out_docker.tar

.PHONY: remote-verify

remote-verify: export WHALE_WATCHER_CONFIG_PATH=./testdata/verify.config.yaml

remote-verify: all oci-export docker-export
	# Verify util signature, not actually perform rule validation
	# Use remote ruleset
	@echo "\n$(BLUE)$(DELIM) Verifying remote ruleset $(DELIM)$(RESET)"
	 ./build/whale-watcher validate https://github.com/coffeemakingtoaster/whale-watcher-target.git $$(pwd)/Dockerfile "./out/out.tar"

.PHONY: local-verify

local-verify: export WHALE_WATCHER_CONFIG_PATH=./testdata/verify.config.yaml

local-verify: all oci-export docker-export
 	# Verify util signature, not actually perform rule validation
	# Use remote ruleset
	@echo "\n$(BLUE)$(DELIM) Verifying local ruleset $(DELIM)$(RESET)"
	./build/whale-watcher validate $$(pwd)/testdata/verify_ruleset.yaml $$(pwd)/Dockerfile "./out/out.tar"

.PHONY: verify
verify: local-verify remote-verify test

.PHONY: docker-verify
docker-verify: docker
	docker run --rm -v $$(pwd)/testdata/verify_ruleset.yaml:/app/verify_ruleset.yaml -v $$(pwd)/Dockerfile:/app/Dockerfile -it whale-watcher:latest "/app/verify_ruleset.yaml" "/app/Dockerfile" "whale-watcher:latest"

.PHONY: install
install: all
	cp ./build/whale-watcher /usr/local/bin/

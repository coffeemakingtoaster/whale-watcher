# Define common paths
BUILD_DIR = ./build
CMD_DIR = ./cmd/whale-watcher
PKG_DIR = ./pkg/runner

# Phony targets declaration
.PHONY: all clean cmd_lib fs_lib os_lib exec

# Ensure the build directory exists
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

cmd_lib: $(BUILD_DIR)/command_util/__init__.py

# Targets for libraries with the build directory as a prerequisite
$(BUILD_DIR)/command_util/__init__.py: $(PKG_DIR)/command_util/command_util.go | $(BUILD_DIR)
	gopy build -output=$(BUILD_DIR)/command_util -vm=python3 $(PKG_DIR)/command_util/

fs_lib: $(BUILD_DIR)/fs_util/__init__.py

$(BUILD_DIR)/fs_util/__init__.py: $(PKG_DIR)/fs_util/fs_util.go | $(BUILD_DIR)
	gopy build -output=$(BUILD_DIR)/fs_util -vm=python3 $(PKG_DIR)/fs_util

os_lib: $(BUILD_DIR)/os_util/__init__.py

$(BUILD_DIR)/os_util/__init__.py: $(PKG_DIR)/os_util/os_util.go | $(BUILD_DIR)
	gopy build -output=$(BUILD_DIR)/os_util -vm=python3 $(PKG_DIR)/os_util

# Target for the executable with the build directory as a prerequisite
exec: $(BUILD_DIR)/whale-watcher

$(BUILD_DIR)/whale-watcher: $(CMD_DIR)/whale-watcher.go | $(BUILD_DIR)
	go build -o $(BUILD_DIR)/whale-watcher $(CMD_DIR)/whale-watcher.go

# Define the main target
all: cmd_lib fs_lib os_lib exec

# Clean target to remove the build directory
clean:
	rm -rf $(BUILD_DIR)

# Run test ruleset that doesn't need a container but performs a basic signature check for the utils
verify: all
	cd build; ./whale-watcher ../_example/verify_ruleset.yaml

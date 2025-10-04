package util_test

import (
	"testing"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/util"
)

func TestUnsafeCheckTrue(t *testing.T) {
	t.Setenv(util.UnsafeModeEnvVar, "true")
	if !util.IsUnsafeMode() {
		t.Error("Unsafe mode mismatch: Expected true Got false")
	}
}

func TestUnsafeCheckFalse(t *testing.T) {
	t.Setenv(util.UnsafeModeEnvVar, "")
	if util.IsUnsafeMode() {
		t.Error("Unsafe mode mismatch: Expected false Got true")
	}
}

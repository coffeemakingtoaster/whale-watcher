package environment_test

import (
	"testing"

	"github.com/coffeemakingtoaster/whale-watcher/internal/environment"
)

func TestUnsafeCheckTrue(t *testing.T) {
	t.Setenv(environment.UnsafeModeEnvVar, "true")
	if !environment.IsUnsafeMode() {
		t.Error("Unsafe mode mismatch: Expected true Got false")
	}
}

func TestUnsafeCheckFalse(t *testing.T) {
	t.Setenv(environment.UnsafeModeEnvVar, "")
	if environment.IsUnsafeMode() {
		t.Error("Unsafe mode mismatch: Expected false Got true")
	}
}

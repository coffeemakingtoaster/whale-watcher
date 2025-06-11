package commandutil_test

import (
	"strings"
	"testing"

	commandutil "iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/runner/command_util"
)

var sampleDockerfile = []string{
	"ARG PREFROMSTATEMENT=true",
	"FROM golang:1.24-bookworm AS build",
	"",
	"WORKDIR /build",
	"# Install deps",
	"RUN apt update && apt install -y python3-pip && \\",
	"python3 -m pip install pybindgen --break-system-packages && \\",
	"go install golang.org/x/tools/cmd/goimports@latest && \\",
	"go install github.com/go-python/gopy@latest",
	"",
	"COPY . .",
	"",
	"# Clean is not necessary here...but better safe than sorry",
	"RUN make clean all verify",
	"",
	"FROM python:3.10-bookworm AS runtime",
	"",
	"WORKDIR /app",
	"",
	"COPY --from=build /build/build/whale-watcher ./whale-watcher",
	"",
	"ENTRYPOINT [\"/app/whale-watcher\"]",
}

func TestStageNodeAt(t *testing.T) {
	cases := []string{
		"build",
		"runtime",
	}
	cu := commandutil.SetupFromContent(sampleDockerfile)

	for i, v := range cases {
		actual := cu.GetStageNodeAt(i)
		if actual.Name != v {
			t.Errorf("Stage node mismatch: Expected %s Got %s", v, actual.ToString())
		}
	}
}

func TestGetAstDepth(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)

	if cu.GetAstDepth() != 2 {
		t.Errorf("Ast depth mismatch: Expected %d Got %d", cu.GetAstDepth(), 2)
	}
}

func TestGetEveryNodeOfInstruction(t *testing.T) {

	cases := map[string]int{
		"FROM":       2,
		"from":       2,
		"ENTRYPOINT": 1,
		"ARG":        1,
	}

	cu := commandutil.SetupFromContent(sampleDockerfile)

	for key, expected := range cases {
		actual := cu.GetEveryNodeOfInstruction(key)
		if len(actual) != expected {
			t.Errorf("Instruction count mismatch: Expected %d Got %d", expected, len(actual))
		}
		actualKey := strings.ToUpper(key)
		for _, node := range actual {
			if node.Instruction() != actualKey {
				t.Errorf("Instruction result mismatch: Expected %s Got %s", actualKey, node.Instruction())
			}
		}
	}
}

package commandutil_test

import (
	"strings"
	"testing"

	"github.com/coffeemakingtoaster/dockerfile-parser/pkg/ast"
	commandutil "github.com/coffeemakingtoaster/whale-watcher/pkg/runner/command_util"
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

func TestGetLastInstructionNodeInStageByCommand(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	index := cu.GetAstDepth() - 1
	command := cu.GetStageNodeAt(index).Instructions[1].Reconstruct()[0]
	valid := cu.GetLastInstructionNodeInStageByCommand(command, index)

	if valid == nil {
		t.Fatalf("Instruction result mismatch: Expected InstructionNode Got Nil (Searched for %s)", command)
	}

	if _, ok := (*valid).(*ast.WorkdirInstructionNode); !ok {
		t.Errorf("Instruction result mismatch: Expected Workdir instruction node Got %v (Searched for %s)", *valid, command)
	}
}

func TestCommandAlwaysHasParamWithCommandNotPresent(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	if !cu.CommandAlwaysHasParam([]string{"invalid"}, "-v") {
		t.Error("Command with param mismatch: Expected true but got false")
	}
}

func TestCommandAlwaysHasParamWithCommandTrueSimple(t *testing.T) {
	alteredDockerfile := append(sampleDockerfile, "RUN curl -f hello")
	cu := commandutil.SetupFromContent(alteredDockerfile)
	if !cu.CommandAlwaysHasParam([]string{"curl"}, "-f") {
		t.Error("Command with param mismatch: Expected true but got false")
	}
}

func TestCommandAlwaysHasParamWithCommandTrueNested(t *testing.T) {
	alteredDockerfile := append(sampleDockerfile, "RUN echo hello && curl -f -x hello && abc def")
	cu := commandutil.SetupFromContent(alteredDockerfile)
	if !cu.CommandAlwaysHasParam([]string{"curl"}, "-f") {
		t.Error("Command with param mismatch: Expected true but got false")
	}
}

func TestCommandAlwaysHasParamWithLongCommandTrue(t *testing.T) {
	alteredDockerfile := append(sampleDockerfile, "RUN echo hello && apt-get install -y vim-btw && abc def")
	cu := commandutil.SetupFromContent(alteredDockerfile)
	if !cu.CommandAlwaysHasParam([]string{"apt-get", "install"}, "-y") {
		t.Error("Command with param mismatch: Expected true but got false")
	}
}

func TestCommandAlwaysHasParamWithCommandFalseSimple(t *testing.T) {
	alteredDockerfile := append(sampleDockerfile, "RUN curl -x hello")
	cu := commandutil.SetupFromContent(alteredDockerfile)
	if cu.CommandAlwaysHasParam([]string{"curl"}, "-f") {
		t.Error("Command with param mismatch: Expected false but got true")
	}
}

func TestCommandAlwaysHasParamWithCommandFalseNested(t *testing.T) {
	alteredDockerfile := append(sampleDockerfile, "RUN echo hello && curl -d -l hello && abc def")
	cu := commandutil.SetupFromContent(alteredDockerfile)
	if cu.CommandAlwaysHasParam([]string{"curl"}, "-f") {
		t.Error("Command with param mismatch: Expected false but got true")
	}
}

func TestCommandAlwaysHasParamWithLongCommandFalse(t *testing.T) {
	alteredDockerfile := append(sampleDockerfile, "RUN echo hello && apt-get install vim-btw && abc def")
	cu := commandutil.SetupFromContent(alteredDockerfile)
	if cu.CommandAlwaysHasParam([]string{"apt-get", "install"}, "-y") {
		t.Error("Command with param mismatch: Expected false but got true")
	}
}

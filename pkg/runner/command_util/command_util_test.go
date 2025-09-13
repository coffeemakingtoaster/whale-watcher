package commandutil_test

import (
	"fmt"
	"reflect"
	"slices"
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
	"(python3 -m pip install pybindgen --break-system-packages && \\",
	"go install golang.org/x/tools/cmd/goimports@latest) && \\",
	"go install github.com/go-python/gopy@latest",
	"",
	"COPY --link . .",
	"",
	"# Clean is not necessary here...but better safe than sorry",
	"RUN make clean all verify",
	"",
	"FROM python:3.10-bookworm AS runtime",
	"",
	"WORKDIR /app",
	"",
	"ENV A=foo B=bar",
	"",
	"ONBUILD RUN echo hello world",
	"COPY --from=build /build/build/whale-watcher ./whale-watcher",
	"EXPOSE 3000",
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
	if !cu.CommandAlwaysHasParam("invalid", "-v") {
		t.Error("Command with param mismatch: Expected true but got false")
	}
}

func TestCommandAlwaysHasParamWithCommandTrueSimple(t *testing.T) {
	alteredDockerfile := append(sampleDockerfile, "RUN curl -f hello")
	cu := commandutil.SetupFromContent(alteredDockerfile)
	if !cu.CommandAlwaysHasParam("curl", "-f") {
		t.Error("Command with param mismatch: Expected true but got false")
	}
}

func TestCommandAlwaysHasParamWithCommandTrueNested(t *testing.T) {
	alteredDockerfile := append(sampleDockerfile, "RUN echo hello && curl -f -x hello && abc def")
	cu := commandutil.SetupFromContent(alteredDockerfile)
	if !cu.CommandAlwaysHasParam("curl", "-f") {
		t.Error("Command with param mismatch: Expected true but got false")
	}
}

func TestCommandAlwaysHasParamWithLongCommandTrue(t *testing.T) {
	alteredDockerfile := append(sampleDockerfile, "RUN echo hello && apt-get install -y vim-btw && abc def")
	cu := commandutil.SetupFromContent(alteredDockerfile)
	if !cu.CommandAlwaysHasParam("apt-get install", "-y") {
		t.Error("Command with param mismatch: Expected true but got false")
	}
}

func TestCommandAlwaysHasParamWithCommandFalseSimple(t *testing.T) {
	alteredDockerfile := append(sampleDockerfile, "RUN curl -x hello")
	cu := commandutil.SetupFromContent(alteredDockerfile)
	if cu.CommandAlwaysHasParam("curl", "-f") {
		t.Error("Command with param mismatch: Expected false but got true")
	}
}

func TestCommandAlwaysHasParamTrueButMixed(t *testing.T) {
	alteredDockerfile := append(sampleDockerfile, "RUN echo hello && apt-get -y install vim-btw && abc def")
	cu := commandutil.SetupFromContent(alteredDockerfile)
	if !cu.CommandAlwaysHasParam("apt-get install", "-y") {
		t.Error("Command with param mismatch: Expected true but got false")
	}
}

func TestCommandAlwaysHasParamWithCommandFalseNested(t *testing.T) {
	alteredDockerfile := append(sampleDockerfile, "RUN echo hello && curl -d -l hello && abc def")
	cu := commandutil.SetupFromContent(alteredDockerfile)
	if cu.CommandAlwaysHasParam("curl", "-f") {
		t.Error("Command with param mismatch: Expected false but got true")
	}
}

func TestCommandAlwaysHasParamWithLongCommandFalse(t *testing.T) {
	alteredDockerfile := append(sampleDockerfile, "RUN echo hello && apt-get install vim-btw && abc def")
	cu := commandutil.SetupFromContent(alteredDockerfile)
	if cu.CommandAlwaysHasParam("apt-get install", "-y") {
		t.Error("Command with param mismatch: Expected false but got true")
	}
}

func TestUsesCommandTrue(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	actual := cu.UsesCommand("python3 -m pip install")

	if !actual {
		t.Errorf("Uses command return mismatch: Exptected %v Got %v", !actual, actual)
	}
}

func TestUsesCommandFalse(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	actual := cu.UsesCommand("imaginary command")
	if actual {
		t.Errorf("Uses command return mismatch: Exptected %v Got %v", !actual, actual)
	}
}

func TestUsesCommandTrueSubSequent(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	actual := cu.UsesCommand("python3 -m pip install")
	if !actual {
		t.Errorf("Uses command return mismatch: Exptected %v Got %v", !actual, actual)
	}
	actual = cu.UsesCommand("make clean")
	if !actual {
		t.Errorf("Uses command return mismatch: Exptected %v Got %v", !actual, actual)
	}
}

func TestGetInstructionNodePropertyStringPropertyExists(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	nodes := cu.GetEveryNodeOfInstruction("WORKDIR")
	actual := cu.GetNodePropertyString(&nodes[0], "Path")
	expected := "/build"
	if actual != expected {
		t.Errorf("Property mismatch: Expected %s Got %s", expected, actual)
	}
}

func TestGetInstructionNodePropertyStringPropertyNotExists(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	nodes := cu.GetEveryNodeOfInstruction("WORKDIR")
	actual := cu.GetNodePropertyString(&nodes[0], "Paths")
	expected := ""
	if actual != expected {
		t.Errorf("Property mismatch: Expected %s Got %s", expected, actual)
	}
}

func TestGetInstructionNodePropertyStringPropertyIsNotString(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	nodes := cu.GetEveryNodeOfInstruction("RUN")
	actual := cu.GetNodePropertyString(&nodes[0], "IsHeredoc")
	expected := ""
	if actual != expected {
		t.Errorf("Property mismatch: Expected %s Got %s", expected, actual)
	}
}

func TestGetInstructionNodePropertyStringListPropertyExists(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	nodes := cu.GetEveryNodeOfInstruction("RUN")
	actual := cu.GetNodePropertyStringList(&nodes[0], "Cmd")
	expected := []string{"apt", "update", "&&", "apt", "install", "-y", "python3-pip", "&&", "(python3", "-m", "pip", "install", "pybindgen", "--break-system-packages", "&&", "go", "install", "golang.org/x/tools/cmd/goimports@latest)", "&&", "go", "install", "github.com/go-python/gopy@latest"}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Property mismatch: Expected %v Got %v", expected, actual)
	}
}

func TestGetInstructionNodePropertyStringPropertyListNotExists(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	nodes := cu.GetEveryNodeOfInstruction("WORKDIR")
	actual := cu.GetNodePropertyStringList(&nodes[0], "Paths")
	expected := []string{}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Property mismatch: Expected %v Got %v", expected, actual)
	}
}

func TestGetInstructionNodePropertyStringPropertyListIsNotStringList(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	nodes := cu.GetEveryNodeOfInstruction("RUN")
	actual := cu.GetNodePropertyStringList(&nodes[0], "IsHeredoc")
	expected := []string{}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Property mismatch: Expected %v Got %v", expected, actual)
	}
}

func TestGetInstructionNodePropertyStringMapKeysPropertyExists(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	nodes := cu.GetEveryNodeOfInstruction("ENV")
	actual := cu.GetNodePropertyStringMapKeys(&nodes[0], "Pairs")
	expected := []string{"A", "B"}
	slices.Sort(expected)
	slices.Sort(actual)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Property mismatch: Expected %v Got %v", expected, actual)
	}
}

func TestGetInstructionNodePropertyStringMapPropertyNotExists(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	nodes := cu.GetEveryNodeOfInstruction("WORKDIR")
	actual := cu.GetNodePropertyStringMapKeys(&nodes[0], "Paths")
	expected := []string{}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Property mismatch: Expected %v Got %v", expected, actual)
	}
}

func TestGetInstructionNodePropertyStringMapPropertyIsNotMap(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	nodes := cu.GetEveryNodeOfInstruction("RUN")
	actual := cu.GetNodePropertyStringMapKeys(&nodes[0], "IsHeredoc")
	expected := []string{}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Property mismatch: Expected %v Got %v", expected, actual)
	}
}

func TestGetInstructionNodePropertyBoolPropertyExists(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	nodes := cu.GetEveryNodeOfInstruction("COPY")
	actual := cu.GetNodePropertyBool(&nodes[0], "Link")
	fmt.Printf("%v %s", nodes[0], nodes[0].Reconstruct())
	expected := true
	if actual != expected {
		t.Errorf("Property mismatch: Expected %v Got %v", expected, actual)
	}
}

func TestGetInstructionNodePropertyBoolPropertyNotExists(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	nodes := cu.GetEveryNodeOfInstruction("WORKDIR")
	actual := cu.GetNodePropertyBool(&nodes[0], "Paths")
	expected := false
	if actual != expected {
		t.Errorf("Property mismatch: Expected %v Got %v", expected, actual)
	}
}

func TestGetInstructionNodePropertyBoolPropertyIsNotMap(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	nodes := cu.GetEveryNodeOfInstruction("RUN")
	actual := cu.GetNodePropertyBool(&nodes[0], "Cmd")
	expected := false
	if actual != expected {
		t.Errorf("Property mismatch: Expected %v Got %v", expected, actual)
	}
}

func TestGetExposeNodePortNumbers(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	nodes := cu.GetEveryNodeOfInstruction("EXPOSE")
	actual := cu.GetExposeNodePortNumbers(&nodes[0])
	expected := []int{3000}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Port slice mismatch: Expected %v Got %v", expected, actual)
	}
}

func TestGetEveryNodeOfInstructionAtLevel(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	expected := []string{"/build", "/app"}
	for i := range 2 {
		nodes := cu.GetEveryNodeOfInstructionAtLevel(i, "WORKDIR")
		for _, node := range nodes {
			workdirNode := node.(*ast.WorkdirInstructionNode)
			if workdirNode.Path != expected[i] {
				t.Errorf("Path mismatch at level %d: Expected %s Got %s", i, expected[i], workdirNode.Path)
			}
		}
	}
}

func TestGetInstructionFromOnbuild(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	nodes := cu.GetEveryNodeOfInstruction("ONBUILD")

	if len(nodes) != 1 {
		t.Errorf("length mismatch: Expected %d Got %d", 1, len(nodes))
	}

	instruction := cu.GetInstructionFromOnbuild(&nodes[0])
	run := instruction.(*ast.RunInstructionNode)
	if !reflect.DeepEqual(run.Cmd, []string{"echo", "hello", "world"}) {
		t.Errorf("Node content mismatch: Expected [echo, hello, world] Got %v", run.Cmd)
	}
}

func TestGetNodeInstructionString(t *testing.T) {
	cu := commandutil.SetupFromContent(sampleDockerfile)
	nodes := cu.GetEveryNodeOfInstruction("RUN")
	for _, node := range nodes {
		if node.Instruction() != cu.GetNodeInstructionString(node) {
			t.Errorf("Instruction string mismatch: Expected %s Got %s", node.Instruction(), cu.GetNodeInstructionString(node))
		}
	}
}

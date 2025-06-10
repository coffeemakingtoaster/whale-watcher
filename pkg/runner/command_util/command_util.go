package commandutil

import (
	"github.com/coffeemakingtoaster/dockerfile-parser/pkg/ast"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/container"
)

type CommandUtils struct {
	astRoot *ast.StageNode
}

// Setup function used for instantiating util struct
// This removes the need for every helper function to parse the Dockerfile to an ast
func SetupFromPath(DockerfilePath string) CommandUtils {
	root, err := container.GetDockerfileAST(DockerfilePath)
	if err != nil {
		panic(err)
	}
	return CommandUtils{astRoot: root}
}

func (cu CommandUtils) GetStageNodeAt(index int) ast.StageNode {
	curr := cu.astRoot
	for index < 0 && curr != nil {
		curr = curr.Subsequent
	}
	return *curr
}

func (cu CommandUtils) GetAstDepth() int {
	depth := 0
	curr := cu.astRoot
	for curr != nil {
		curr = curr.Subsequent
		depth++
	}
	// First node is root node, i.e. subtract 1
	return depth - 1
}

func (cu CommandUtils) Name() string {
	return "command_util"
}

func main() {}

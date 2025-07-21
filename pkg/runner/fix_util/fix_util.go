package fixutil

import (
	"os"
	"strings"

	"github.com/coffeemakingtoaster/dockerfile-parser/pkg/ast"
	"github.com/rs/zerolog/log"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/container"
)

type FixUtils struct {
	astRoot *ast.StageNode
	path    string
}

// Setup function used for instantiating util struct
// This removes the need for every helper function to parse the Dockerfile to an ast
func SetupFromPath(DockerfilePath string) FixUtils {
	root, err := container.GetDockerfileAST(DockerfilePath)
	if err != nil {
		panic(err)
	}
	return FixUtils{astRoot: root, path: DockerfilePath}
}

func SetupFromContent(DockerfileContent []string) FixUtils {
	root, err := container.GetDockerfileInputAST(DockerfileContent)
	if err != nil {
		panic(err)
	}
	return FixUtils{astRoot: root}
}

func (cu *FixUtils) AddRunInstruction(index int, command string) {
	curr := cu.astRoot
	ind := -1
	for ind < index {
		curr = curr.Subsequent
		ind++
	}
	curr.Instructions = append(curr.Instructions, &ast.RunInstructionNode{
		Cmd: []string{command},
	})
}

func (cu *FixUtils) SetUser(user string) {
	curr := cu.astRoot
	for curr.Subsequent != nil {
		curr = curr.Subsequent
	}
	last := curr.Instructions[len(curr.Instructions)-1]
	curr.Instructions[len(curr.Instructions)-1] = &ast.UserInstructionNode{User: user}
	curr.Instructions = append(curr.Instructions, last)
}

func (fu *FixUtils) Finish() {
	newContent := fu.astRoot.Reconstruct()
	if fu.path == "" {
		return
	}

	dockerFile, err := os.OpenFile(fu.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Error().Err(err)
		return
	}

	data := strings.Join(newContent, "\n")

	dockerFile.Write([]byte(data))

	dockerFile.Close()
}

func main() {}

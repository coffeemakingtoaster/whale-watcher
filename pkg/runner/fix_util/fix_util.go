package fixutil

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/coffeemakingtoaster/dockerfile-parser/pkg/ast"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/container"
	"github.com/rs/zerolog/log"
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

// Find the last place where a run instruction contains the search string and append the command
// Mainly meant for adding cleanup commands
func (cu *FixUtils) AppendRunInstructionWithMatch(search, command string) bool {
	layer_index := -1
	instruction_index := -1
	curr_index := 0
	curr := cu.astRoot
	for curr != nil {
		for instruction_count, instruction := range curr.Instructions {
			switch run := instruction.(type) {
			case *ast.RunInstructionNode:
				for i := range run.Cmd {
					if strings.Contains(run.Cmd[i], search) {
						instruction_index = instruction_count
						layer_index = curr_index
					}
				}
			}
		}
		curr_index++
		curr = curr.Subsequent
	}
	if layer_index == -1 {
		return false
	}

	stage := cu.getStage(layer_index)

	instr := stage.Instructions[instruction_index].(*ast.RunInstructionNode)
	instr.Cmd = append(instr.Cmd, []string{"&&", command}...)

	return true
}

func (cu *FixUtils) getStage(index int) *ast.StageNode {
	curr := cu.astRoot
	for index > 0 {
		curr = curr.Subsequent
		index--
	}
	return curr
}

func (cu *FixUtils) AddRunInstruction(command string) {
	curr := cu.astRoot
	for curr.Subsequent != nil {
		curr = curr.Subsequent
	}
	last := curr.Instructions[len(curr.Instructions)-1]
	curr.Instructions[len(curr.Instructions)-1] = &ast.RunInstructionNode{
		Cmd: []string{command},
	}
	curr.Instructions = append(curr.Instructions, last)
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

func (cu *FixUtils) CreateUser(user string) {
	curr := cu.astRoot
	for curr.Subsequent != nil {
		curr = curr.Subsequent
	}
	cu.AddRunInstruction(fmt.Sprintf("groupadd -r %s && useradd --no-log-init -r -g %s %s", user, user, user))
	cu.SetUser(user)
}

// This is very inefficient
// To make this faster the parser should likely change
// if only the maintainer would have the time
// TODO: Add a fix counterpart to this
func (fu *FixUtils) EnsureCommandAlwaysHasParam(command, param string) {
	curr := fu.astRoot
	for curr != nil {
		for i := range curr.Instructions {
			runNode, ok := curr.Instructions[i].(*ast.RunInstructionNode)
			if !ok {
				continue
			}
			fu.ensureCommandInNodeAlwaysHasParam(runNode, command, param)
		}
		curr = curr.Subsequent
	}
}

func (fu *FixUtils) ensureCommandInNodeAlwaysHasParam(node *ast.RunInstructionNode, command, param string) {
	pointer := 0
	for pointer < len(node.Cmd) {
		cmd := node.Cmd[pointer]
		if cmd == command {
			pointer++
			cmdPointer := pointer
			for pointer < len(node.Cmd) {
				cmd := node.Cmd[pointer]
				if strings.Contains(cmd, param) {
					break
				}
				// is command block ended/Run instruction end without finding wanted param?
				if cmd == "&&" {
					node.Cmd = slices.Insert(node.Cmd, cmdPointer, param)
					break
				}
				if pointer == len(node.Cmd)-1 {
					node.Cmd = slices.Insert(node.Cmd, cmdPointer, param)
					return
				}
				pointer++
			}
		}
		pointer++
	}
}

func (fu *FixUtils) GetReconstruct() []string {
	return fu.astRoot.Reconstruct()
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

	log.Info().Str("data", data).Send()

	dockerFile.Write([]byte(data))

	dockerFile.Close()
}

func main() {}

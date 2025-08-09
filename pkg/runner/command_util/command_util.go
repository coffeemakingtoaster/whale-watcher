package commandutil

import (
	"fmt"
	"strings"

	"github.com/coffeemakingtoaster/dockerfile-parser/pkg/ast"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/container"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/util"
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

func SetupFromContent(DockerfileContent []string) CommandUtils {
	root, err := container.GetDockerfileInputAST(DockerfileContent)
	if err != nil {
		panic(err)
	}
	return CommandUtils{astRoot: root}
}

func (cu *CommandUtils) GetStageNodeAt(index int) *ast.StageNode {
	curr := cu.astRoot.Subsequent
	for index > 0 && curr != nil {
		curr = curr.Subsequent
		index--
	}
	return curr
}

// This is a temporary solution...I need a better way to expose the data
func (cu *CommandUtils) GetStageName(sn ast.StageNode) string {
	return sn.Name
}

func (cu *CommandUtils) GetEveryNodeOfInstruction(wantedInstruction string) []ast.Node {
	wantedInstruction = strings.ToUpper(wantedInstruction)
	res := []ast.Node{}
	currNode := cu.astRoot
	for currNode != nil {
		// empty root node has image of ""
		if currNode.Instruction() == wantedInstruction && currNode.Image != "" {
			res = append(res, currNode)
			currNode = currNode.Subsequent
			continue
		}
		for _, instructionNode := range currNode.Instructions {
			if instructionNode.Instruction() == wantedInstruction {
				res = append(res, instructionNode)
			}
		}
		currNode = currNode.Subsequent
	}
	return res
}

func (cu *CommandUtils) GetAstDepth() int {
	depth := 0
	curr := cu.astRoot
	for curr != nil {
		curr = curr.Subsequent
		depth++
	}
	// First node is root node, i.e. subtract 1
	return depth - 1
}

// Given a command, try to get the last instruction from the ast that matches the command
// This approximates and COULD fail
// This does not support parser annotations
// This is the fast version -> A slow version may just lex + parse the command
func (cu *CommandUtils) GetLastInstructionNodeInStageByCommand(command string, stage int) *ast.InstructionNode {
	var result *ast.InstructionNode //start with nil
	if stage >= cu.GetAstDepth() {
		return result
	}

	currentNode := cu.astRoot

	for range stage + 1 {
		currentNode = currentNode.Subsequent
	}

	for _, instruction := range currentNode.Instructions {
		fmt.Println(instruction.Reconstruct())
		if command != instruction.Reconstruct()[0] {
			continue
		}
		result = &instruction
	}
	return result
}

func (cu *CommandUtils) UsesSubstringAnywhere(pattern string) bool {
	curr := cu.astRoot
	for curr != nil {
		if strings.Contains(strings.Join(curr.Reconstruct(), "\n"), pattern) {
			return true
		}
		curr = curr.Subsequent
	}
	return false
}

// Check if command is used in Dockerfile
// TODO: Add way to check stage
func (cu *CommandUtils) UsesCommand(command string) bool {
	search := util.NewSliceSearch(strings.Split(command, " "))
	runNodes := cu.GetEveryNodeOfInstruction("RUN")
	for _, n := range runNodes {
		node, ok := n.(*ast.RunInstructionNode)
		if !ok {
			continue
		}
		for i := range node.Cmd {
			if search.Match(node.Cmd[i]) {
				return true
			}
		}
		search.Reset()
	}
	return false
}

// This is very inefficient
// To make this faster the parser should likely change
// if only the maintainer would have the time
// This is unable to detect parameters weaved into command (apt-get -y install will not detect the -y)
func (cu *CommandUtils) CommandAlwaysHasParam(rawCommand string, param string) bool {
	command := strings.Split(rawCommand, " ")
	nodes := cu.GetEveryNodeOfInstruction("RUN")
	for _, node := range nodes {
		search := util.NewSliceSearch(command)
		runNode, ok := node.(*ast.RunInstructionNode)
		if !ok {
			panic("Conversion error")
		}
		pointer := 0
		for pointer < len(runNode.Cmd) {
			cmd := runNode.Cmd[pointer]
			if search.Match(cmd) {
				pointer++
				for pointer < len(runNode.Cmd) {
					cmd := runNode.Cmd[pointer]
					if strings.Contains(cmd, param) {
						break
					}
					// is command block ended/Run instruction end without finding wanted param?
					if cmd == "&&" || pointer == len(runNode.Cmd)-1 {
						return false
					}
					pointer++
				}
				search.Reset()
			}
			pointer++
		}
	}
	return true
}

func (cu *CommandUtils) Name() string {
	return "command_util"
}

func main() {}

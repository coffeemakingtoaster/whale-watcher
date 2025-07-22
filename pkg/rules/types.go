package rules

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/runner"
)

var allowedCategories = []string{"negative", "positive"}
var allowedScopes = []string{"output", "buildtime"}
var allowedTargets = []string{"command", "os", "fs"}

type ViolationInfo struct {
	Details string
	Fix     string
}

type RuleSet struct {
	Name       string  `yaml:"name"`
	Rules      []*Rule `yaml:"rules"`
	tmpDirPath string
}

type Rule struct {
	Scope           string `yaml:"scope"`
	Category        string `yaml:"category"`
	Instruction     string `yaml:"instruction"`
	Description     string `yaml:"description"`
	LongDescription string `yaml:"long_description"`
	Id              string `yaml:"id"`
	Target          string `yaml:"target"`
	Runner          runner.Runner
	FixInstruction  string `yaml:"fix_instruction"`
}

func (r *Rule) AddRunner() error {
	var err error
	r.Runner, err = runner.NewPythonRunner(r.Target)
	return err
}

func (r *Rule) Validate(ociTarPath, dockerFilepath, dockerTarPath string) (bool, ViolationInfo) {
	err := r.Runner.Run(runner.TemplateData{DockerfilePath: dockerFilepath, OciImage: ociTarPath, DockerImage: dockerTarPath}, r.Instruction)
	if err != nil {
		return false, ViolationInfo{Details: err.Error()}
	}
	return true, ViolationInfo{}
}

// Cleanup if this was loaded from git
func (rs *RuleSet) Close() {
	if len(rs.tmpDirPath) > 0 {
		os.RemoveAll(rs.tmpDirPath)
	}
}

func (r *Rule) Verify() error {
	// TODO: Add instruction verify as soon as helper format is clear
	if len(r.Id) == 0 {
		return errors.New("No id set for rule")
	}
	r.Category = strings.ToLower(r.Category)
	if err := isInAllowed(r.Category, allowedCategories); err != nil {
		return fmt.Errorf("Category: %s", err.Error())
	}
	r.Scope = strings.ToLower(r.Scope)
	if err := isInAllowed(r.Scope, allowedScopes); err != nil {
		return fmt.Errorf("Scope: %s", err.Error())
	}
	r.Target = strings.ToLower(r.Target)
	if err := isInAllowed(r.Target, allowedTargets); err != nil {
		return fmt.Errorf("Target: %s", err.Error())
	}
	return nil
}

func (r *Rule) PerformFix() error {
	if r.FixInstruction == "" {
		return errors.New("No fixinstruction present")
	}
	r.Runner.RunFix(r.FixInstruction)
	return nil
}

func isInAllowed(value string, allowList []string) error {
	if !slices.Contains(allowList, value) {
		return errors.New(fmt.Sprintf("Invalid value %s (Allowed: %+q)", value, allowList))
	}
	return nil
}

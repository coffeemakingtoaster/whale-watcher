package rules

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/runner"
	"github.com/rs/zerolog/log"
)

var allowedCategories = []string{"negative", "positive"}
var allowedTargets = []string{"command", "os", "fs"}

type ViolationInfo struct {
	Details string
	Fix     string
}

type RuleSet struct {
	Name       string   `yaml:"name"`
	Include    []string `yaml:"include"`
	Rules      []*Rule  `yaml:"rules"`
	tmpDirPath string
	ids        map[string]int
	targetList map[string]bool
}

type Rule struct {
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

func (r *Rule) GetUtilLevel() int {
	switch r.Target {
	case "command":
		return runner.COMMAND_UTIL_LEVEL
	case "fs":
		return runner.FS_UTIL_LEVEL
	case "os":
		return runner.OS_UTIL_LEVEL
	default:
		log.Warn().Str("target", r.Target).Msgf("Unknown target, falling back to os")
		return runner.OS_UTIL_LEVEL
	}
}

func (r *Rule) Validate(ociTarPath, dockerFilepath, dockerTarPath string) (bool, ViolationInfo) {
	err := r.Runner.Run(runner.TemplateData{DockerfilePath: dockerFilepath, OciImage: ociTarPath, DockerImage: dockerTarPath}, r.Instruction, r.GetUtilLevel())
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

func (rs *RuleSet) updateIdList() {
	rs.ids = make(map[string]int)
	for i, rule := range rs.Rules {
		rs.ids[rule.Id] = i
	}
}

// Considers currently target allowlist in config
func (rs *RuleSet) GetHighestTarget() string {
	for _, target := range []string{"os", "fs"} {
		if val, ok := rs.targetList[target]; ok && val && config.AllowsTarget(target) {
			return target
		}
	}
	return "command"
}

// Take all rules fromt he weaker set where the current set does not have a rule yet
// identified via ID
func (rs *RuleSet) Swallow(weakerSet RuleSet) {
	if len(rs.Rules) != len(rs.ids) {
		rs.updateIdList()
	}
	if rs.targetList == nil {
		rs.targetList = make(map[string]bool)
	}
	for _, rule := range weakerSet.Rules {
		if _, ok := rs.ids[rule.Id]; !ok {
			rs.Rules = append(rs.Rules, rule)
			rs.targetList[rule.Target] = true
		}
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

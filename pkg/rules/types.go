package rules

import (
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/runner"
)

type ViolationInfo struct {
	Details string
	Fix     string
}

type RuleSet struct {
	Name  string  `yaml:"name"`
	Rules []*Rule `yaml:"rules"`
}

type Rule struct {
	Scope       string `yaml:"scope"`
	Category    string `yaml:"category"`
	Instruction string `yaml:"instruction"`
	Description string `yaml:"description"`
	Id          string `yaml:"id"`
	Target      string `yaml:"target"`
	Runner      runner.Runner
}

func (r *Rule) AddRunner() error {
	var err error
	r.Runner, err = runner.NewPythonRunner(r.Target)
	return err
}

func (r *Rule) Validate(imageName, dockerFilepath string) (bool, ViolationInfo) {
	err := r.Runner.Run(runner.TemplateData{DockerfilePath: dockerFilepath, Image: imageName}, r.Instruction)
	if err != nil {
		return false, ViolationInfo{Details: err.Error()}
	}
	return true, ViolationInfo{}
}

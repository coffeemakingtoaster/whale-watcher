package runner

import (
	"github.com/coffeemakingtoaster/whale-watcher/pkg/rules"
)

type RunnerResult struct {
	Autofix bool
	Success bool
}

type Runner interface {
	Run(ruleSet rules.RuleSet, ociTarPath, dockerFilepath, dockerTarPath string) (map[string]RunnerResult, error)
}

package command

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog"
	"iteragit.iteratec.de/max.herkenhoff/whale-watcher/pkg/fetcher"
)

func getContext(args []string) (*RunContext, error) {
	if len(args) < 1 {
		return nil, errors.New("No valid command passed. Use help for more detail!")
	}
	runContext := RunContext{Instruction: args[0]}
	switch runContext.Instruction {
	case "validate":
		err := runContext.parseVerify(args[1:])
		if err != nil {
			return nil, err
		}
	case "bic":
	case "help":
	case "docs":
		err := runContext.parseDocs(args[1:])
		if err != nil {
			return nil, err
		}
		// Only show errors now, otherwise interference with documentation
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		return nil, errors.New(fmt.Sprintf("Unknown command: %s. Use help for more detail!", runContext.Instruction))
	}
	return &runContext, nil
}

func (rc *RunContext) parseVerify(args []string) error {
	if len(args) < 1 {
		return errors.New("Not enough arguments for validate. Needs <ruleset location> (other values are set via config)")
	}
	// Load from config
	rc.DockerFile, rc.OCITarballPath, rc.DockerTarballPath = fetcher.FetchContainerFiles()
	rc.RuleSetEntrypoint = args[0]
	return nil
}

func (rc *RunContext) parseDocs(args []string) error {
	if len(args) < 1 {
		return errors.New("Not enough arguments for docs. Needs <ruleset location>.")
	}
	rc.RuleSetEntrypoint = args[0]
	return nil
}

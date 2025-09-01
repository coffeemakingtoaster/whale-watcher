package command

import (
	"fmt"

	"github.com/coffeemakingtoaster/whale-watcher/internal/display"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/adapters"
	baseimagecache "github.com/coffeemakingtoaster/whale-watcher/pkg/base_image_cache"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/base_image_cache/ingester"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/container"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/rules"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/runner"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/validator"
	"github.com/rs/zerolog/log"
)

type RunContext struct {
	OCITarballPath    string
	DockerTarballPath string
	DockerFile        string
	RuleSetEntrypoint string
	Instruction       string
	ExportHTML        bool
}

var helpText = `
Whale watcher!
Valid commands:
	- help -> its this one :)
	- validate <ruleset> <dockerfile> <oci image tarball> -> validate the ruleset against the given container artifacts
	- docs <ruleset> -> Serve the ruleset documentation as a website. Pass --export to output and index.html instead
	- bic -> build base image cache
	`

func Run(args []string) int {
	runContext, err := getContext(args)
	if err != nil {
		panic(err)
	}
	if runContext.Instruction == "help" {
		fmt.Println(helpText)
		return 0
	}
	if runContext.Instruction == "bic" {
		cfg := config.GetConfig()
		if len(cfg.BaseImageCache.BaseImages) == 0 {
			log.Warn().Msg("No base images listed. Nothing to do...")
			return 1
		}
		for _, img := range cfg.BaseImageCache.BaseImages {
			ingester.IngestImage(img)
		}
		return 0
	}
	ruleSet, err := rules.LoadRuleset(runContext.RuleSetEntrypoint)
	if err != nil {
		panic(err)
	}
	log.Info().Msgf("Loaded %d rules!", len(ruleSet.Rules))
	if runContext.Instruction == "docs" {
		display.ServeRules(ruleSet, runContext.ExportHTML)
		return 0
	}
	// Get ref to prevent directory cleanup
	ref := runner.GetReferencingWorkingDirectoryInstance()
	defer ref.ForceFree()

	violations := getViolations(runContext, ruleSet)

	// should a pr be created?
	if config.ShouldInteractWithVSC() {
		err = adapters.CreatePRForFixes(violations, ref.GetAbsolutePath("./Dockerfile"))
		if err != nil {
			log.Error().Err(err).Msg("Failed to create PR for changes/fixes")
		}
	} else {
		log.Info().Msg("No git context, no interaction with VSC platform needed")
	}
	err = recommendBaseImage(ref, violations)

	if err != nil {
		return 1
	}

	if violations.ViolationCount > 0 {
		return 1
	}
	return 0
}

func getViolations(runContext *RunContext, ruleSet rules.RuleSet) validator.Violations {
	// TODO: These paths are passed down way to far without any validation
	violations := validator.ValidateRuleset(ruleSet, runContext.OCITarballPath, runContext.DockerFile, runContext.DockerTarballPath)
	log.Info().Msgf("Total: %d Violations: %d Fixable: %d", violations.CheckedCount, violations.ViolationCount, violations.FixableCount)
	for _, violation := range violations.Violations {
		log.Warn().Str("ruleId", violation.RuleId).Str("problem", violation.Description).Send()
	}
	return violations
}

func recommendBaseImage(ref *runner.RunnerWorkingDirectory, violations validator.Violations) error {
	cfg := config.GetConfig()
	// This is fine in of itself -> not configured
	if len(cfg.BaseImageCache.BaseImages)+len(cfg.BaseImageCache.CacheLocation) == 0 {
		return nil
	}
	baseImageCache := baseimagecache.NewBaseImageCache()
	loadedImage, err := container.ContainerImageFromOCITar(ref.GetAbsolutePath("./out.tar"))
	if err != nil {
		log.Warn().Err(err).Msg("Could not parse oci tar")
		return err
	}
	if loadedImage.GetBaseImage() != "" {
		log.Info().Str("base image", loadedImage.GetBaseImage()).Msg("Already uses known base image")
		return nil
	}
	closestBaseImage, err := baseImageCache.GetClosestDependencyImage(loadedImage.GetPackageList())
	if err != nil || len(closestBaseImage) == 0 {
		log.Warn().Err(err).Msg("Could not determine closest base image")
		return err
	}
	if config.ShouldInteractWithVSC() && violations.ViolationCount > 0 {
		log.Debug().Msg("Trying to update PR with base image hint")
		adapter, err := adapters.GetAdapterForRepository(cfg.Target.RepositoryURL)
		if err != nil {
			log.Warn().Err(err).Msg("Could not set for adapter")
			return err
		}
		description := violations.BuildDescriptionMarkdown()
		description += fmt.Sprintf("\n⚠️ Recommended Base Image: `%s` ⚠️\n", closestBaseImage)
		err = adapter.UpdatePullRequest("", description)
		if err != nil {
			log.Warn().Err(err).Msg("Could not update PR")
			return err
		}
	}
	log.Info().Str("base image", closestBaseImage).Msg("Found fitting base image!")
	return nil
}

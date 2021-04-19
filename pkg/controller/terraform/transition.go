package terraform

import (
	"errors"
	"fmt"
	"github.com/xmidt-org/carousel/pkg/controller"
	"github.com/xmidt-org/carousel/pkg/model"
	"github.com/xmidt-org/carousel/pkg/runner"
	"strings"
)

var (
	errTaintHostFailure = errors.New("failed to taint host")
)

// tTransition is a terraform specific implementation of Transition
type tTransition struct {
	// applyRunBuilder is a helper function that generates a Runnable for a given Step
	config           model.BinaryConfig
	transitionConfig TerraformTransitionConfig
}

func (t *tTransition) CreateApply(target model.ClusterState, step model.Step) runner.Runnable {
	cmdArgs := []string{
		"apply", "--auto-approve",
	}

	for _, color := range model.ValidColors {
		cmdArgs = append(cmdArgs,
			"-var", fmt.Sprintf("version%sCount=%d", strings.Title(color.String()), step[color]),
		)
		cmdArgs = append(cmdArgs,
			"-var", fmt.Sprintf("version%s=%s", strings.Title(color.String()), target[color].Version.String()),
		)
	}

	for _, elem := range t.transitionConfig.Args {
		cmdArgs = append(cmdArgs, "-var", fmt.Sprintf("%s=%s", elem.Key, elem.Value))
	}

	runConfig := runner.Options{
		ShowOutput:        t.transitionConfig.AttachStdOut,
		SuppressErrOutput: !t.transitionConfig.AttachStdErr,
	}
	r := runner.NewCMDRunner(t.config.WorkingDirectory, t.config.Binary, runConfig, cmdArgs...)
	r = runner.AddEnvironment(r, "TF_VAR_", t.config.PrivateArgs)
	r = runner.AddEnvironment(r, "", t.config.Environment)

	return r
}

type TerraformTransitionConfig struct {
	Args         []model.ValuePair
	AttachStdOut bool
	AttachStdErr bool
}

// BuildTransitioner builds a terraform specific controller.ApplyBuilder.
func BuildTransitioner(config model.BinaryConfig, transitionConfig TerraformTransitionConfig) controller.ApplyBuilder {
	return &tTransition{
		config:           config,
		transitionConfig: transitionConfig,
	}
}

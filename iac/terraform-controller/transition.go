package terraform_controller

import (
	"errors"
	"fmt"
	"github.com/xmidt-org/carousel/iac"
	"github.com/xmidt-org/carousel/model"
	"github.com/xmidt-org/carousel/runner"
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

	r := runner.NewCMDRunner(t.config.WorkingDirectory, t.config.Binary, false, t.transitionConfig.AttachStdOut, t.transitionConfig.AttachStdErr, cmdArgs...)
	r = runner.AddEnvironment(r, "TF_VAR_", t.config.PrivateArgs)
	r = runner.AddEnvironment(r, "", t.config.Environment)

	return r
}

type TerraformTransitionConfig struct {
	Args         []model.ValuePair
	AttachStdOut bool
	AttachStdErr bool
}

// BuildTransitioner builds a terraform specific iac.ApplyBuilder.
func BuildTransitioner(config model.BinaryConfig, transitionConfig TerraformTransitionConfig) iac.ApplyBuilder {
	return &tTransition{
		config:           config,
		transitionConfig: transitionConfig,
	}
}

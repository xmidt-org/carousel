package carousel

import (
	"errors"
	"fmt"
	"github.com/blang/semver/v4"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"strings"
	"sync"
)

var (
	errGoalStateFailure = errors.New("failed to establish goal state")
	errTaintHostFailure = errors.New("failed to taint host")
)

// tTransition is a terraform specific implementation of Transition
type tTransition struct {
	// applyRunBuilder is a helper function that generates a Runnable for a given Step
	applyRunBuilder   func(target ClusterState, step Step) Runnable
	validatorResource HostValidator
	tainter           Tainter
	CG                ClusterGetter
	getGoalState      GoalStateFunc
	stepOptions       []StepOptions
	logger            log.Logger
	dryRun            bool
}

// handleRun runs a Runnable until an unrecoverable error occurs or all host created are valid.
func (t *tTransition) handleRun(applyrunner Runnable, currentHosts map[string]bool, applyGroup Color) error {
	t.logger.Log(level.Key(), level.DebugValue(), "runner", applyrunner.String())
	// aka. terraform apply step
	out, err := applyrunner.Output()
	if err != nil {
		// todo:// try some other handler logic
		return RunnableError{
			Output:    out,
			ResultErr: fmt.Errorf("%w: with runnable %s", err, applyrunner.String()),
		}
	}
	rerun := false

	// get the resulting cluster.
	newCluster, err := t.CG.GetCluster()
	if err != nil {
		return err
	}

	hostsToCheck := make([]string, 0)

	// check each new host to see if its valid.
	for _, host := range newCluster[applyGroup].Hosts {
		if !currentHosts[host] {
			hostsToCheck = append(hostsToCheck, host)
		}
	}
	hostToCheckCount := len(hostsToCheck)
	errChan := make(chan error, hostToCheckCount)
	wg := new(sync.WaitGroup)
	wg.Add(hostToCheckCount)
	for _, host := range hostsToCheck {
		go func() {
			if !t.validatorResource(host) {
				t.logger.Log(level.Key(), level.DebugValue(), "msg", "check failed", "host", host)

				err = t.tainter.TaintHost(host)
				if err != nil {
					errChan <- err
				}

				// re run command and checks stuff	t.logger.Log(level.Key(), level.DebugValue(), "msg", cc.AsClusterState().String())
				rerun = true
			} else {
				currentHosts[host] = true
			}
			wg.Done()
		}()
	}
	wg.Wait()
	close(errChan)
	// if a host is not valid we have to rerun the step.
	if rerun {
		return t.handleRun(applyrunner, currentHosts, applyGroup)
	}

	var taintingErrors Errors
	for taintErr := range errChan {
		if taintErr != nil {
			taintingErrors = append(taintingErrors, taintErr)
		}
	}
	if len(taintingErrors.Errors()) == 0 {
		return nil
	}
	return taintingErrors
}

func (t *tTransition) Transition(nodeCount int, version semver.Version) error {
	// Get the current cluster.
	cc, err := t.CG.GetCluster()
	t.logger.Log(level.Key(), level.DebugValue(), "msg", cc.AsClusterState().String())

	if err != nil {
		return fmt.Errorf("%w: %v", errGetClusterFailure, err)
	}
	// Determine the goal state.
	goalCluster, err := t.getGoalState(cc.AsClusterState(), nodeCount, version)
	if err != nil {
		return fmt.Errorf("%w: %v", errGoalStateFailure, err)
	}
	currentGroup, _ := cc.AsClusterState().Group()

	// Build the steps to get to goal
	steps := CreateSteps(cc.AsClusterState(), goalCluster, t.stepOptions...)

	return t.ApplySteps(cc, currentGroup, steps, goalCluster)
}

func (t *tTransition) Resume(startingColor Color, steps []Step, goalCluster ClusterState) error {
	// Get the current cluster.
	cc, err := t.CG.GetCluster()
	t.logger.Log(level.Key(), level.DebugValue(), "msg", cc.AsClusterState().String())
	if err != nil {
		return fmt.Errorf("%w: %v", errGetClusterFailure, err)
	}

	return t.ApplySteps(cc, startingColor, steps, goalCluster)
}

// apply steps will apply the given steps to get to a a cluster state
// the first step must match the current cluster
func (t *tTransition) ApplySteps(currentCluster Cluster, currentGroup Color, steps []Step, goalCluster ClusterState) error {
	if !currentCluster.AsClusterState().IsEmpty() {
		if len(steps) < 2 {
			return errors.New("len of steps must be greater than 2")
		}

		// check first step matches currentCluster
		if !currentCluster.AsClusterState().EqualStep(steps[0]) {
			return errors.New("current cluster doesn't match first step")
		}
	}

	// check last step matches goal state
	if !goalCluster.EqualStep(steps[len(steps)-1]) {
		return errors.New("end cluster doesn't match last step")
	}

	// for each step apply it.
	// on error take current and future steps and return them in addition to the error.
	currentHosts := map[string]bool{}
	for _, host := range currentCluster[currentGroup.Other()].Hosts {
		currentHosts[host] = true
	}
	// run each step
	for index, step := range steps {
		applyRunner := t.applyRunBuilder(goalCluster, step)
		if t.dryRun {
			t.logger.Log(level.Key(), level.InfoValue(), "msg", applyRunner.String())
			continue
		}
		err := t.handleRun(applyRunner, currentHosts, currentGroup.Other())
		if err != nil {
			// TODO: better error handling
			return StepError{
				Cause:              err,
				TODO:               steps[index:],
				OriginalCluster:    currentCluster,
				StartingColorGroup: currentGroup,
				GoalClusterState:   goalCluster,
			}
		}
		t.logger.Log(level.Key(), level.InfoValue(), "msg", fmt.Sprintf("completed step: blue with %d nodes and green with %d nodes", step[Blue], step[Green]))
	}
	return nil
}

type TerraformTransitionConfig struct {
	Args         []ValuePair
	DryRun       bool
	AttachStdOut bool
	AttachStdErr bool
}

// BuildTransitioner builds a terraform specific Transition.
func BuildTransitioner(config BinaryConfig, logger log.Logger, validator HostValidator, tainter Tainter, getter ClusterGetter, state GoalStateFunc, transitionConfig TerraformTransitionConfig, stepOptions ...StepOptions) Transition {
	return &tTransition{
		applyRunBuilder: func(target ClusterState, step Step) Runnable {
			cmdArgs := []string{
				"apply", "--auto-approve",
			}

			for _, color := range ValidColors {
				cmdArgs = append(cmdArgs,
					"-var", fmt.Sprintf("version%sCount=%d", strings.Title(color.String()), step[color]),
				)
				cmdArgs = append(cmdArgs,
					"-var", fmt.Sprintf("version%s=%s", strings.Title(color.String()), target[color].Version.String()),
				)
			}

			for _, elem := range transitionConfig.Args {
				cmdArgs = append(cmdArgs, "-var", fmt.Sprintf("%s=%s", elem.Key, elem.Value))
			}

			r := NewCMDRunner(config.WorkingDirectory, config.Binary, false, transitionConfig.AttachStdOut, transitionConfig.AttachStdErr, cmdArgs...)
			r = AddEnvironment(r, "TF_VAR_", config.PrivateArgs)
			r = AddEnvironment(r, "", config.Environment)

			return r
		},
		validatorResource: validator,
		tainter:           tainter,
		CG:                getter,
		getGoalState:      state,
		logger:            logger,
		stepOptions:       stepOptions,
		dryRun:            transitionConfig.DryRun,
	}
}

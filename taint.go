package carousel

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

var errNoSuchResource = errors.New("resource not found, could be an outdated terraform https://github.com/hashicorp/terraform/pull/22467")

type tTaint struct {
	graphCluster ClusterGraph
	// taintRunnerBuilder is a helper function that generates a Runnable to taint a given resource dependency.
	taintRunnerBuilder func(key string) Runnable
}

func (t *tTaint) TaintResources(resources []string) error {
	for _, key := range resources {
		tRunner := t.taintRunnerBuilder(key)
		_, err := tRunner.Output()
		// check if we succeeded in tainting the resource, its possible that a resource is un-taint-able.
		if err != nil {
			validError := t.checkRecoverable(err)
			if validError != nil {
				return fmt.Errorf("%w: tainting failed with runner: %s", err, tRunner.String())
			}
			// TODO: leverage output from runner to create better debug statements
		}
	}
	return nil
}

// checkRecoverable checks the error to see if it is recoverable.
func (t *tTaint) checkRecoverable(err error) error {
	if err == nil {
		return nil
	}
	errString := err.Error()
	if exitError, ok := err.(*exec.ExitError); ok {
		errString = string(exitError.Stderr)
	}
	if exitError, ok := err.(ExitError); ok {
		errString = string(exitError.CapturedErrorOutput)
	}

	if strings.Contains(strings.ToLower(errString), "cannot be tainted") {
		return nil
	}
	if strings.Contains(strings.ToLower(errString), "no such resource instance") {
		return errNoSuchResource
	}
	return err
}

func (t *tTaint) TaintHost(hostname string) error {
	// get the dependencies for a given host
	resources, err := t.graphCluster.GetResourcesForHost(hostname)
	if err != nil {
		return fmt.Errorf("%w: %v for host %s", errTaintHostFailure, err, hostname)
	}
	// taint all the dependencies
	return t.TaintResources(resources)
}

// BuildTaintHostRunner builds a terraform specific Tainter.
func BuildTaintHostRunner(graphCluster ClusterGraph, config BinaryConfig) Tainter {
	return &tTaint{
		graphCluster: graphCluster,
		taintRunnerBuilder: func(key string) Runnable {
			return NewCMDRunner(config.WorkingDirectory, config.Binary, false, false, false, "taint", key)
		},
	}
}

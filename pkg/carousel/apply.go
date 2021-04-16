package carousel

import (
	"errors"
	"fmt"
	"github.com/go-kit/kit/log/level"
	"github.com/xmidt-org/carousel/pkg/model"
	"github.com/xmidt-org/carousel/pkg/runner"
	"sync"
)

// transition will apply the given steps to get to a cluster state to its goal state.
// the first step must match the current cluster.
func (c Carousel) transition(currentCluster model.Cluster, currentGroup model.Color, steps []model.Step, goalCluster model.ClusterState) error {
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
		applyRunner := c.controller.CreateApply(goalCluster, step)
		if c.config.DryRun {
			c.ui.Info(applyRunner.String())
			continue
		}
		err := c.handleRun(applyRunner, currentHosts, currentGroup.Other())
		if err != nil {
			// TODO: better error handling
			return model.StepError{
				Cause:              err,
				TODO:               steps[index:],
				OriginalCluster:    currentCluster,
				StartingColorGroup: currentGroup,
				GoalClusterState:   goalCluster,
			}
		}
		c.ui.Info(fmt.Sprintf("completed step: blue with %d nodes and green with %d nodes", step[model.Blue], step[model.Green]))
	}
	return nil
}

// handleRun runs a Runnable until an unrecoverable error occurs or all host created are valid.
func (c Carousel) handleRun(applyRunner runner.Runnable, currHost map[string]bool, applyGroup model.Color) error {
	level.Debug(c.logger).Log("runner", applyRunner.String())

	// aka. terraform apply step
	out, err := applyRunner.Output()
	if err != nil {
		// todo:// try some other handler logic
		return model.RunnableError{
			Output:    out,
			ResultErr: fmt.Errorf("%w: with runnable %s", err, applyRunner.String()),
		}
	}

	// get the resulting cluster.
	newCluster, err := c.controller.GetCluster()
	if err != nil {
		return err
	}

	hostsToCheck := make([]string, 0)

	// check each new host to see if its valid.
	for _, host := range newCluster[applyGroup].Hosts {
		if !currHost[host] {
			hostsToCheck = append(hostsToCheck, host)
		}
	}
	hostToCheckCount := len(hostsToCheck)
	errChan := make(chan error, hostToCheckCount)
	reRunChan := make(chan bool, hostToCheckCount)
	wg := new(sync.WaitGroup)
	wg.Add(hostToCheckCount)
	for _, host := range hostsToCheck {
		go c.checkHost(host, errChan, reRunChan, currHost, wg)
	}
	wg.Wait()
	close(errChan)
	close(reRunChan)
	// if a host is not valid we have to rerun the step.
	for range reRunChan {
		return c.handleRun(applyRunner, currHost, applyGroup)
	}

	var taintingErrors model.Errors
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

func (c Carousel) checkHost(hostname string, errChan chan<- error, rerun chan<- bool, currHost map[string]bool, wg *sync.WaitGroup) {
	if !c.config.Validate(hostname) {
		level.Debug(c.logger).Log("msg", "check failed", "host", hostname)

		err := c.controller.TaintHost(hostname)
		if err != nil {
			errChan <- err
		}

		// re run command and checks stuff	t.logger.Log(level.Key(), level.DebugValue(), "msg", cc.AsClusterState().String())
		rerun <- true
	} else {
		currHost[hostname] = true
	}
	wg.Done()
}

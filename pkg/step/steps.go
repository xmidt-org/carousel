package step

import (
	"github.com/xmidt-org/carousel/pkg/model"
)

// Generator is a function that returns the Step(s) to build from a ClusterState to a target ClusterState.
type Generator func(currentCluster model.ClusterState, targetCluster model.ClusterState, stepOptions ...StepOptions) []model.Step

type options struct {
	batchSize  int
	skipFirstN int
}

// StepOptions represents options for creating building steps
type StepOptions func(*options)

// WithBatchSize configures how to build the steps for going from a current ClusterState to a goal ClusterState.
// Size MUST be greater than 0, the default value is 1.
// If >1 then each step will change by no more than the value set.
func WithBatchSize(size int) StepOptions {
	return func(o *options) {
		if size <= 0 {
			o.batchSize = 1
		}
		o.batchSize = size
	}
}

// WithSkipFirstN configures how to build the steps for going from a current ClusterState to a goal ClusterState.
// n MUST not be negative, the default value is 0.
// For example if set to 2, then the cluster will never have 2 or fewer servers in a Color Group.
func WithSkipFirstN(n int) StepOptions {
	return func(o *options) {
		if n < 0 {
			o.skipFirstN = 0
		}
		o.skipFirstN = n
	}
}

// CreateSteps will generate the Blue, Green steps to switch from a current ClusterState to a target ClusterState.
//
// Starting point for rollback
//
// Phase 1 (only applicable if new cluster is larger than current one)
// We want to build the extra machines we will need for the new cluster
//
// Phase 2: We can now alternate between adding a node to the new cluster and removing one from the old one
// By the time we're done, we should have all nodes needed for the new cluster
//
// Phase 3: Remove all remaining nodes from the old cluster
func CreateSteps(currentCluster model.ClusterState, targetCluster model.ClusterState, stepOptions ...StepOptions) []model.Step {
	// Nothing to do
	if currentCluster.IsEmpty() && targetCluster.IsEmpty() {
		return []model.Step{AsStep(targetCluster)}
	}
	options := &options{
		batchSize:  1,
		skipFirstN: 0,
	}
	for _, updateFunc := range stepOptions {
		updateFunc(options)
	}

	buildColor, _ := targetCluster.Group()
	if buildColor == model.Unknown {
		buildColor = model.ValidColors[0]
	}
	return append([]model.Step{AsStep(currentCluster)}, generateSteps(currentCluster, targetCluster, options, buildColor, true, []model.Step{})...)
}

// generateSteps is a tail recursive call for building steps with the create step prepended to the list.
func generateSteps(currentCluster model.ClusterState, targetCluster model.ClusterState, options *options, group model.Color, addNodes bool, steps []model.Step) []model.Step {
	// BaseCase
	if currentCluster.EqualNodeCount(targetCluster) {
		return steps
	}
	var nextState model.ClusterState

	var (
		currentNodeCount = currentCluster[group].Count
		targetNodeCount  = targetCluster[group].Count
	)
	if currentCluster[group].Count < targetCluster[group].Count && addNodes {
		nextState = currentCluster.AddNodes(group, addAndSkip(currentNodeCount, targetNodeCount, options.batchSize, options.skipFirstN))
	} else if currentCluster[group].Count > targetCluster[group].Count && !addNodes {
		nextState = currentCluster.AddNodes(group, -minusAndSkip(currentNodeCount, targetNodeCount, options.batchSize, options.skipFirstN))
	} else { // Can't add or remove nodes in group anymore
		var (
			otherCurrentNodeCount = currentCluster[group.Other()].Count
			otherTargetNodeCount  = targetCluster[group.Other()].Count
		)
		if currentCluster[group.Other()].Count < targetCluster[group.Other()].Count {
			nextState = currentCluster.AddNodes(group.Other(), addAndSkip(otherCurrentNodeCount, otherTargetNodeCount, options.batchSize, options.skipFirstN))
		} else if currentCluster[group.Other()].Count > targetCluster[group.Other()].Count {
			nextState = currentCluster.AddNodes(group.Other(), -minusAndSkip(otherCurrentNodeCount, otherTargetNodeCount, options.batchSize, options.skipFirstN))
		} else if currentCluster[group].Count < targetCluster[group].Count {
			nextState = currentCluster.AddNodes(group, addAndSkip(currentNodeCount, targetNodeCount, options.batchSize, options.skipFirstN))
		} else if currentCluster[group].Count > targetCluster[group].Count {
			nextState = currentCluster.AddNodes(group, -minusAndSkip(currentNodeCount, targetNodeCount, options.batchSize, options.skipFirstN))
		} else {
			panic("next state not created")
		}
	}

	return append([]model.Step{AsStep(nextState)}, generateSteps(nextState, targetCluster, options, group.Other(), !addNodes, steps)...)
}

func addAndSkip(currentNodeCount int, targetNodeCount int, batchSize int, skipfirstN int) int {
	if currentNodeCount > skipfirstN && currentNodeCount+batchSize >= targetNodeCount {
		return targetNodeCount - currentNodeCount
	} else if currentNodeCount+batchSize <= skipfirstN {
		return skipfirstN + 1
	} else { // currentNodeCount > skipfirstN  && currentNodeCount + batchSize < targetNodeCount
		return batchSize
	}
}

func minusAndSkip(currentNodeCount int, targetNodeCount int, batchSize int, skipfirstN int) int {
	if currentNodeCount-batchSize <= skipfirstN {
		return currentNodeCount - targetNodeCount
	} else {
		return batchSize
	}
}

// AsStep creates a Step struct from a ClusterState
// In other words a Step will result in a ClusterState.
// This is a helper function to make life easier.
func AsStep(cs model.ClusterState) model.Step {
	step := model.Step{}
	for color, group := range cs {
		step[color] = group.Count
	}
	return step
}

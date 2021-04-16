package goal

import (
	"errors"
	"fmt"
	"github.com/blang/semver/v4"
	"github.com/xmidt-org/carousel/pkg/model"
)

var (
	ErrDetermineGroupFailure = errors.New("failed to determine current group")
)

// BuildEndState is a GoalStateFunc that provides a simple 2 Color Group switch.
func BuildEndState(current model.ClusterState, nodeCount int, version semver.Version) (model.ClusterState, error) {
	currentGroup, err := current.Group()
	if err != nil {
		return model.NewClusterState(), fmt.Errorf("%w: %v", ErrDetermineGroupFailure, err)
	}
	goal := current.Clone()
	goal[currentGroup] = model.ClusterGroupState{
		Count:   0,
		Version: current[currentGroup].Version, // TODO; should the version stay?
	}
	goal[currentGroup.Other()] = model.ClusterGroupState{
		Count:   nodeCount,
		Version: version,
	}
	return goal, nil
}

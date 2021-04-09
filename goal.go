package carousel

import (
	"errors"
	"fmt"
	"github.com/blang/semver/v4"
)

var (
	ErrDetermineGroupFailure = errors.New("failed to determine current group")
)

// DefaultGoalState is a GoalStateFunc that provides a simple 2 Color Group switch.
func DefaultGoalState(current ClusterState, nodeCount int, version semver.Version) (ClusterState, error) {
	currentGroup, err := current.Group()
	if err != nil {
		return NewClusterState(), fmt.Errorf("%w: %v", ErrDetermineGroupFailure, err)
	}
	goal := current.Clone()
	goal[currentGroup] = ClusterGroupState{
		Count:   0,
		Version: current[currentGroup].Version, // TODO; should the version stay?
	}
	goal[currentGroup.Other()] = ClusterGroupState{
		Count:   nodeCount,
		Version: version,
	}
	return goal, nil
}

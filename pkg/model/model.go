package model

import (
	"errors"
	"fmt"
	"github.com/blang/semver/v4"
	"strings"
)

var (
	errNotCleanClusterState = errors.New("cluster is not in a clean state")
)

type ClusterGroup struct {
	Hosts   []string       `json:"hosts"`
	Version semver.Version `json:"version"`
}

type ClusterGroupState struct {
	Count   int            `json:"count"`
	Version semver.Version `json:"version"`
}

// Cluster is a representation of how Cluster is by the hostnames and version deployed.
// Note the Color Groups MUST be the same in both Hosts and Version.
type Cluster map[Color]ClusterGroup

// NewCluster creates a new Cluster with all the ValidColors set to an empty version with no Nodes.
func NewCluster() Cluster {
	cs := Cluster{}
	for _, color := range ValidColors {
		cs[color] = ClusterGroup{
			Version: semver.Version{},
			Hosts:   make([]string, 0),
		}
	}
	return cs
}

// Step is the equivalent to the arguments of a single terraform command.
// (aka -var versionBlueCount=10 -var versionGreenCount=5)
type Step map[Color]int

func (s Step) String() string {
	str := ""
	for color, count := range s {
		str += fmt.Sprintf("%s:%d ", strings.Title(color.String()), count)
	}
	return str
}

// ClusterState is a simplified representation of Cluster where
// NodeCount equals the number of servers in the Color Group.
type ClusterState map[Color]ClusterGroupState

func (cs ClusterState) String() string {
	str := ""
	for color, group := range cs {
		str += fmt.Sprintf("%s@%s:%d ", strings.Title(color.String()), group.Version, group.Count)
	}
	return str
}

// NewClusterState creates a new ClusterState with all the ValidColors set to an empty version with no Nodes.
func NewClusterState() ClusterState {
	cs := ClusterState{}
	for _, color := range ValidColors {
		cs[color] = ClusterGroupState{}
	}
	return cs
}

// IsEmpty returns true if all the Color Groups have zero nodes.
func (cs ClusterState) IsEmpty() bool {
	for _, group := range cs {
		if group.Count != 0 {
			return false
		}
	}
	return true
}

// IsCleanState returns true if only one Color Group has nodes.
func (cs ClusterState) IsCleanState() bool {
	groupsWithNodes := 0
	for _, group := range cs {
		if group.Count != 0 {
			groupsWithNodes++
		}
	}
	// if no nodes in cluster, still a "clean" cluster
	return groupsWithNodes <= 1
}

// Group returns the Color Group if the ClusterState is in a clean state,  (see IsCleanState)
// otherwise returns the Unknown Color Group.
func (cs ClusterState) Group() (Color, error) {
	if !cs.IsCleanState() {
		return Unknown, errNotCleanClusterState
	}
	for color, group := range cs {
		if group.Count != 0 {
			return color, nil
		}
	}
	return ValidColors[0], nil
}

// EqualNodeCount returns true if both structs have the same NodeCount map
func (cs ClusterState) EqualNodeCount(other ClusterState) bool {
	if len(cs) != len(other) {
		return false
	}
	for color := range cs {
		if cs[color].Count != other[color].Count {
			return false
		}
	}
	return true
}
func (cs ClusterState) EqualStep(step Step) bool {
	if len(cs) != len(step) {
		return false
	}
	for color := range cs {
		if cs[color].Count != step[color] {
			return false
		}
	}
	return true
}

func (cs ClusterState) Clone() ClusterState {
	newCS := NewClusterState()
	for color := range cs {
		newCS[color] = cs[color]
	}
	return newCS
}

// AddNodes creates a new ClusterState with the addition of nodes to the Color Group.
func (cs ClusterState) AddNodes(group Color, count int) ClusterState {
	newCS := cs.Clone()
	newCS[group] = ClusterGroupState{
		Count:   newCS[group].Count + count,
		Version: newCS[group].Version,
	}
	return newCS
}

// AsClusterState returns a ClusterState struct from a given Cluster
func (c Cluster) AsClusterState() ClusterState {
	cs := NewClusterState()
	for color, group := range c {
		cs[color] = ClusterGroupState{
			Count:   len(group.Hosts),
			Version: group.Version,
		}
	}
	return cs
}

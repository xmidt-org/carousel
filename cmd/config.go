package main

import (
	"github.com/xmidt-org/carousel/model"
)

// RolloutConfig specifies the options for transitioning the cluster to the new state.
type RolloutConfig struct {
	// SkipFirstN will make it so the cluster never has <N number of nodes in a group
	// For example if set to 2, then the cluster will never have 2 or fewer servers in a Color Group.
	SkipFirstN int
	// BatchSize configures how many nodes can be batched at once.
	// If >1 then each step will change by no more than the value set.
	BatchSize int
}

// Config provides the configuration to the carousel binary.
type Config struct {
	// Workspace the terraform workspace to use. If empty, the current workspace will be used.
	Workspace     string
	BinaryConfig  model.BinaryConfig
	RolloutConfig RolloutConfig
}

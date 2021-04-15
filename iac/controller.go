package iac

import (
	"errors"
	"github.com/xmidt-org/carousel/model"
	"github.com/xmidt-org/carousel/runner"
)

var (
	ErrGetClusterFailure = errors.New("failed to get cluster state")
	ErrGoalStateFailure  = errors.New("failed to establish goal state")
)

// WorkspaceSelecter is something that can change the workspace.
type WorkspaceSelecter interface {
	// SelectWorkspace changes the workspace used.
	// If an empty string is supplied, the current workspace is used.
	SelectWorkspace(workspace string) error
}

// WorkspaceSelecter is something that can change get the current Cluster.
type ClusterGetter interface {
	// GetCluster returns the cluster or an error.
	GetCluster() (model.Cluster, error)
}

// Tainter is something that can mark something as bad.
type Tainter interface {
	// TaintResources will mark the given resources as bad or returns an error.
	TaintResources(resources []string) error
	// TaintHost will mark a Host as bad or returns an error.
	TaintHost(hostname string) error
}

type ApplyBuilder interface {
	CreateApply(target model.ClusterState, step model.Step) runner.Runnable
}

// ClusterGraph is something that can get the resource dependencies of a specific host.
type ClusterGraph interface {
	// GetResourcesForHost returns the resource dependencies or an error given a specific host
	GetResourcesForHost(hostname string) ([]string, error)
}

type Controller interface {
	WorkspaceSelecter
	ClusterGetter
	Tainter
	ApplyBuilder
}

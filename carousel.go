package carousel

import (
	"github.com/blang/semver/v4"
)

// ValuePair represents a key value pair and is a workaround for viper converting all keys to lower case.
// https://github.com/spf13/viper/issues/371
type ValuePair struct {
	// Key is the case sensitive argument or environment name.
	Key string

	// Value is the associated value.
	Value string
}

// BinaryConfig represent the configuration in order to execute a terraform command.
type BinaryConfig struct {
	// Binary is the path to the binary to be Ran.
	// This will will search the PATH for which binary to use.
	// For more information refer to https://golang.org/pkg/os/exec/#LookPath
	// If empty `terraform` will be used.
	Binary string

	// Args are optional arguments to be supplied to the binary. Note this will show up in plain text.
	Args []ValuePair

	// PrivateArgs are arguments that will supplied to the binary via Environment Variables with the Prefix TF_VAR_.
	// Refer to https://www.terraform.io/docs/cli/config/environment-variables.html#tf_var_name for more information.
	PrivateArgs []ValuePair

	// Environment is additional environment variables to give the binary to run on top of the current environment.
	Environment []ValuePair

	// WorkingDirectory is the working directory to run the specified Binary.
	// If empty the current directory will be used.
	WorkingDirectory string
}

// SelectWorkspace is something that can change the workspace.
type SelectWorkspace interface {
	// SelectWorkspace changes the workspace used.
	// If an empty string is supplied, the current workspace is used.
	SelectWorkspace(workspace string) error
}

// SelectWorkspace is something that can change get the current Cluster.
type ClusterGetter interface {
	// GetCluster returns the cluster or an error.
	GetCluster() (Cluster, error)
}

// GoalStateFunc creates a Goal ClusterState given a current Cluster.
type GoalStateFunc func(current ClusterState, nodeCount int, version semver.Version) (ClusterState, error)

// ClusterGraph is something that can get the resource dependencies of a specific host.
type ClusterGraph interface {
	// GetResourcesForHost returns the resource dependencies or an error given a specific host
	GetResourcesForHost(hostname string) ([]string, error)
}

// Tainter is something that can mark something as bad.
type Tainter interface {
	// TaintResources will mark the given resources as bad or returns an error.
	TaintResources(resources []string) error
	// TaintHost will mark a Host as bad or returns an error.
	TaintHost(hostname string) error
}

// HostValidator is a function that Checks if a Host is bad or good.
type HostValidator func(fqdn string) bool

func AsHostValidator(f func(fqdn string) bool) HostValidator {
	return f
}

// StepGenerator is a function that returns the Step(s) to build from a ClusterState to a target ClusterState.
type StepGenerator func(currentCluster ClusterState, targetCluster ClusterState, stepOptions ...StepOptions) []Step

// Transition is something that can transition to a new ClusterState.
type Transition interface {
	// Transition handles the transition to a new ClusterState and returns an error if a problem occurs.
	Transition(nodeCount int, version semver.Version) error

	// Resume handles transition with specific steps from a starting color to a goal cluster with specified steps.
	Resume(startingColor Color, steps []Step, goalCluster ClusterState) error
}

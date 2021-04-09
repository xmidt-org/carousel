package carousel

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/blang/semver/v4"
	"github.com/hashicorp/terraform/states/statefile"
	"github.com/zclconf/go-cty/cty"
)

var (
	errFailedToGetData   = errors.New("failed to pull state")
	errBuildStateFailure = errors.New("failed to build terraform state")
)

type tState struct {
	stateRunner Runnable
}

func (t *tState) GetCluster() (Cluster, error) {
	c := NewCluster()
	data, err := t.stateRunner.Output()
	if err != nil {
		return c, fmt.Errorf("%w: %v", errFailedToGetData, err)
	}

	r := bytes.NewReader(data)

	sf, err := statefile.Read(r)
	if err != nil {
		return c, fmt.Errorf("%w: %v", errBuildStateFailure, err)
	}
	s := sf.State
	if s == nil || s.Empty() {
		return c, nil
	}

	// Get the hostnames for each group
	for _, color := range ValidColors {
		var (
			hosts   []string
			version semver.Version
		)
		hostnamesKey := fmt.Sprintf("%sHostnames", color)
		if hostnamesElem, ok := s.RootModule().OutputValues[hostnamesKey]; ok && hostnamesElem != nil {
			hostsSlice := hostnamesElem.Value.AsValueSlice()
			hosts = make([]string, len(hostsSlice))
			for i, val := range hostsSlice {
				if !val.IsNull() {
					if val.Type() == cty.String {
						// TODO: if not a string, return an error.
						hosts[i] = val.AsString()
					}
				}
			}
		}

		versionKey := fmt.Sprintf("%sVersion", color)
		if versionElem, ok := s.RootModule().OutputValues[versionKey]; ok && versionElem != nil {
			version, err = semver.Parse(versionElem.Value.AsString())
			if err != nil {
				return c, fmt.Errorf("%w: %s %s", err, color, versionElem.Value.AsString())
			}
		}
		c[color] = ClusterGroup{
			Hosts:   hosts,
			Version: version,
		}
	}
	return c, nil
}

// BuildStateDeterminer builds a terraform specific ClusterGetter.
func BuildStateDeterminer(config BinaryConfig) ClusterGetter {
	return &tState{
		stateRunner: NewCMDRunner(config.WorkingDirectory, config.Binary, false, false, false, "state", "pull"),
	}
}

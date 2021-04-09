package carousel

import (
	"errors"
	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

var emptyCluster = NewCluster()

func TestGetCluster(t *testing.T) {
	tests := []struct {
		name            string
		expectedCluster Cluster
		runner          Runnable
		expectedErr     error
	}{
		{
			name: "empty_output",
			runner: simplerunnable{
				Name: "testRunner",
				Data: nil,
			},
			expectedCluster: emptyCluster,
			expectedErr:     errFailedToGetData,
		},
		{
			name: "empty_json",
			runner: simplerunnable{
				Name: "testRunner",
				Data: []byte(`{}`),
			},
			expectedCluster: emptyCluster,
			expectedErr:     errBuildStateFailure,
		},
		{
			name: "empty_state",
			runner: simplerunnable{
				Name: "testRunner",
				Data: []byte(emptyState),
			},
			expectedCluster: emptyCluster,
		},
		{
			name: "clean_state",
			runner: simplerunnable{
				Name: "testRunner",
				Data: []byte(cleanState),
			},
			expectedCluster: Cluster{
				Blue: ClusterGroup{
					Hosts:   []string{},
					Version: semver.MustParse("0.10.0"),
				},
				Green: ClusterGroup{
					Hosts:   []string{"carousel-demo-ffdbb6.example.com", "carousel-demo-ea9412.example.com"},
					Version: semver.MustParse("0.10.0"),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			stateGetter := tState{stateRunner: test.runner}
			cluster, err := stateGetter.GetCluster()
			if test.expectedErr != nil {
				assert.True(errors.Is(err, test.expectedErr))
			} else {
				assert.NoError(err)
			}
			assert.Equal(test.expectedCluster, cluster)
		})
	}
}

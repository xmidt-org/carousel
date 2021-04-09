package carousel

import (
	"errors"
	"fmt"
	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultGoal(t *testing.T) {
	tests := []struct {
		name            string
		sourceCluster   ClusterState
		expectedCluster ClusterState
		nodeCount       int
		version         semver.Version
		expectedErr     error
	}{
		{
			name:          "empty_cluster",
			sourceCluster: NewClusterState(),
			expectedCluster: ClusterState{
				Blue: ClusterGroupState{},
				Green: ClusterGroupState{
					Count:   3,
					Version: semver.MustParse("0.1.1"),
				},
			},
			nodeCount: 3,
			version:   semver.MustParse("0.1.1"),
		},
		{
			name: "clear_cluster",
			sourceCluster: ClusterState{
				Blue: ClusterGroupState{},
				Green: ClusterGroupState{
					Count:   3,
					Version: semver.MustParse("0.1.1"),
				},
			},
			expectedCluster: ClusterState{
				Blue: ClusterGroupState{},
				Green: ClusterGroupState{
					Count:   0,
					Version: semver.MustParse("0.1.1"),
				},
			},
			nodeCount: 0,
			version:   semver.MustParse("0.0.0"),
		},
		{
			name: "increaseCluster",
			sourceCluster: ClusterState{
				Blue: ClusterGroupState{},
				Green: ClusterGroupState{
					Count:   2,
					Version: semver.MustParse("0.0.1"),
				},
			},
			expectedCluster: ClusterState{
				Green: ClusterGroupState{
					Version: semver.MustParse("0.0.1"),
				},
				Blue: ClusterGroupState{
					Count:   3,
					Version: semver.MustParse("0.1.1"),
				},
			},
			nodeCount: 3,
			version:   semver.MustParse("0.1.1"),
		},
		{
			name: "decreaseCluster",
			sourceCluster: ClusterState{
				Green: ClusterGroupState{},
				Blue: ClusterGroupState{
					Count:   10,
					Version: semver.MustParse("0.13.4"),
				},
			},
			expectedCluster: ClusterState{
				Blue: ClusterGroupState{
					Version: semver.MustParse("0.13.4"),
				},
				Green: ClusterGroupState{
					Count:   5,
					Version: semver.MustParse("1.0.0"),
				},
			},
			nodeCount: 5,
			version:   semver.MustParse("1.0.0"),
		},
		{
			name: "sameCluster",
			sourceCluster: ClusterState{
				Blue: ClusterGroupState{},
				Green: ClusterGroupState{
					Count:   3,
					Version: semver.MustParse("0.1.1"),
				},
			},
			expectedCluster: ClusterState{
				Green: ClusterGroupState{
					Version: semver.MustParse("0.1.1"),
				},
				Blue: ClusterGroupState{
					Count:   3,
					Version: semver.MustParse("0.1.1"),
				},
			},
			nodeCount: 3,
			version:   semver.MustParse("0.1.1"),
		},
		{
			name: "invalidStart",
			sourceCluster: ClusterState{
				Green: ClusterGroupState{
					Count:   1,
					Version: semver.Version{},
				},
				Blue: ClusterGroupState{
					Count:   3,
					Version: semver.MustParse("0.1.1"),
				},
			},
			expectedCluster: NewClusterState(),
			nodeCount:       3,
			version:         semver.MustParse("0.1.1"),
			expectedErr:     ErrDetermineGroupFailure,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			actualCluster, err := DefaultGoalState(test.sourceCluster, test.nodeCount, test.version)
			if test.expectedErr != nil {
				fmt.Println(err, test.expectedErr)
				assert.True(errors.Is(err, test.expectedErr))
			} else {
				assert.NoError(err)
			}
			if !assert.Equal(test.expectedCluster, actualCluster) {
				t.Log(test.expectedCluster, actualCluster)
			}
		})
	}
}

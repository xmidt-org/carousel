package goal

import (
	"errors"
	"fmt"
	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/assert"
	"github.com/xmidt-org/carousel/pkg/model"
	"testing"
)

func TestDefaultGoal(t *testing.T) {
	tests := []struct {
		name            string
		sourceCluster   model.ClusterState
		expectedCluster model.ClusterState
		nodeCount       int
		version         semver.Version
		expectedErr     error
	}{
		{
			name:          "empty_cluster",
			sourceCluster: model.NewClusterState(),
			expectedCluster: model.ClusterState{
				model.Blue: model.ClusterGroupState{},
				model.Green: model.ClusterGroupState{
					Count:   3,
					Version: semver.MustParse("0.1.1"),
				},
			},
			nodeCount: 3,
			version:   semver.MustParse("0.1.1"),
		},
		{
			name: "clear_cluster",
			sourceCluster: model.ClusterState{
				model.Blue: model.ClusterGroupState{},
				model.Green: model.ClusterGroupState{
					Count:   3,
					Version: semver.MustParse("0.1.1"),
				},
			},
			expectedCluster: model.ClusterState{
				model.Blue: model.ClusterGroupState{},
				model.Green: model.ClusterGroupState{
					Count:   0,
					Version: semver.MustParse("0.1.1"),
				},
			},
			nodeCount: 0,
			version:   semver.MustParse("0.0.0"),
		},
		{
			name: "increaseCluster",
			sourceCluster: model.ClusterState{
				model.Blue: model.ClusterGroupState{},
				model.Green: model.ClusterGroupState{
					Count:   2,
					Version: semver.MustParse("0.0.1"),
				},
			},
			expectedCluster: model.ClusterState{
				model.Green: model.ClusterGroupState{
					Version: semver.MustParse("0.0.1"),
				},
				model.Blue: model.ClusterGroupState{
					Count:   3,
					Version: semver.MustParse("0.1.1"),
				},
			},
			nodeCount: 3,
			version:   semver.MustParse("0.1.1"),
		},
		{
			name: "decreaseCluster",
			sourceCluster: model.ClusterState{
				model.Green: model.ClusterGroupState{},
				model.Blue: model.ClusterGroupState{
					Count:   10,
					Version: semver.MustParse("0.13.4"),
				},
			},
			expectedCluster: model.ClusterState{
				model.Blue: model.ClusterGroupState{
					Version: semver.MustParse("0.13.4"),
				},
				model.Green: model.ClusterGroupState{
					Count:   5,
					Version: semver.MustParse("1.0.0"),
				},
			},
			nodeCount: 5,
			version:   semver.MustParse("1.0.0"),
		},
		{
			name: "sameCluster",
			sourceCluster: model.ClusterState{
				model.Blue: model.ClusterGroupState{},
				model.Green: model.ClusterGroupState{
					Count:   3,
					Version: semver.MustParse("0.1.1"),
				},
			},
			expectedCluster: model.ClusterState{
				model.Green: model.ClusterGroupState{
					Version: semver.MustParse("0.1.1"),
				},
				model.Blue: model.ClusterGroupState{
					Count:   3,
					Version: semver.MustParse("0.1.1"),
				},
			},
			nodeCount: 3,
			version:   semver.MustParse("0.1.1"),
		},
		{
			name: "invalidStart",
			sourceCluster: model.ClusterState{
				model.Green: model.ClusterGroupState{
					Count:   1,
					Version: semver.Version{},
				},
				model.Blue: model.ClusterGroupState{
					Count:   3,
					Version: semver.MustParse("0.1.1"),
				},
			},
			expectedCluster: model.NewClusterState(),
			nodeCount:       3,
			version:         semver.MustParse("0.1.1"),
			expectedErr:     ErrDetermineGroupFailure,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			actualCluster, err := BuildEndState(test.sourceCluster, test.nodeCount, test.version)
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

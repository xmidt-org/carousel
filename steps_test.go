package carousel

import (
	"github.com/blang/semver/v4"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSteps(t *testing.T) {
	tests := []struct {
		name          string
		sourceCluster ClusterState
		targetCluster ClusterState
		options       []StepOptions
		expectedSteps []Step
	}{
		{
			name:          "empty_cluster",
			sourceCluster: NewClusterState(),
			targetCluster: ClusterState{
				Blue: ClusterGroupState{
					Count:   3,
					Version: semver.Version{},
				},
				Green: ClusterGroupState{},
			},
			expectedSteps: []Step{
				{Blue: 0, Green: 0},
				{Blue: 1, Green: 0},
				{Blue: 2, Green: 0},
				{Blue: 3, Green: 0},
			},
		},
		{
			name: "clear_cluster",
			sourceCluster: ClusterState{
				Green: ClusterGroupState{
					Count:   3,
					Version: semver.Version{},
				},
				Blue: ClusterGroupState{}, // TODO test empty color set
			},
			targetCluster: NewClusterState(), // TODO:  Test empty struct
			expectedSteps: []Step{
				{Blue: 0, Green: 3},
				{Blue: 0, Green: 2},
				{Blue: 0, Green: 1},
				{Blue: 0, Green: 0},
			},
		},
		{
			name: "batch_by_3",
			sourceCluster: ClusterState{
				Green: ClusterGroupState{
					Count:   3,
					Version: semver.Version{},
				},
				Blue: ClusterGroupState{}, // TODO test empty color set
			},
			targetCluster: ClusterState{
				Blue: ClusterGroupState{
					Count:   6,
					Version: semver.Version{},
				},
				Green: ClusterGroupState{},
			},
			options: []StepOptions{
				WithBatchSize(3),
			},
			expectedSteps: []Step{
				{Blue: 0, Green: 3},
				{Blue: 3, Green: 3},
				{Blue: 3, Green: 0},
				{Blue: 6, Green: 0},
			},
		},
		{
			name: "batch_by_3_off_ending",
			sourceCluster: ClusterState{
				Green: ClusterGroupState{
					Count:   3,
					Version: semver.Version{},
				},
				Blue: ClusterGroupState{},
			},
			targetCluster: ClusterState{
				Blue: ClusterGroupState{
					Count:   7,
					Version: semver.Version{},
				},
				Green: ClusterGroupState{},
			},
			options: []StepOptions{
				WithBatchSize(3),
			},
			expectedSteps: []Step{
				{Blue: 0, Green: 3},
				{Blue: 3, Green: 3},
				{Blue: 3, Green: 0},
				{Blue: 6, Green: 0},
				{Blue: 7, Green: 0},
			},
		},
		{
			name: "switch_group_same_count_skip_1",
			sourceCluster: ClusterState{
				Blue: ClusterGroupState{
					Count:   4,
					Version: semver.Version{},
				},
				Green: ClusterGroupState{},
			},
			targetCluster: ClusterState{
				Green: ClusterGroupState{
					Count:   4,
					Version: semver.Version{},
				},
				Blue: ClusterGroupState{},
			},
			options: []StepOptions{
				WithSkipFirstN(1),
			},
			expectedSteps: []Step{
				{Green: 0, Blue: 4},
				{Green: 2, Blue: 4},
				{Green: 2, Blue: 3},
				{Green: 3, Blue: 3},
				{Green: 3, Blue: 2},
				{Green: 4, Blue: 2},
				{Green: 4, Blue: 0},
			},
		},
		{
			name: "smaller_target_cluster",
			sourceCluster: ClusterState{
				Green: ClusterGroupState{
					Count:   5,
					Version: semver.Version{},
				},
				Blue: ClusterGroupState{},
			},
			targetCluster: ClusterState{
				Blue: ClusterGroupState{
					Count:   3,
					Version: semver.Version{},
				},
				Green: ClusterGroupState{},
			},
			expectedSteps: []Step{
				{Blue: 0, Green: 5},
				{Blue: 1, Green: 5},
				{Blue: 1, Green: 4},
				{Blue: 2, Green: 4},
				{Blue: 2, Green: 3},
				{Blue: 3, Green: 3},
				{Blue: 3, Green: 2},
				{Blue: 3, Green: 1},
				{Blue: 3, Green: 0},
			},
		},
		{
			name: "empty_skip_3",
			sourceCluster: ClusterState{
				Blue: ClusterGroupState{
					Count:   4,
					Version: semver.Version{},
				},
				Green: ClusterGroupState{},
			},
			targetCluster: NewClusterState(),
			options: []StepOptions{
				WithSkipFirstN(2),
			},
			expectedSteps: []Step{
				{Green: 0, Blue: 4},
				{Green: 0, Blue: 3},
				{Green: 0, Blue: 0},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			actualSteps := CreateSteps(test.sourceCluster, test.targetCluster, test.options...)
			assert.Equal(test.expectedSteps, actualSteps)
		})
	}
}

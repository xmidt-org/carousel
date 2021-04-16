package step

import (
	"github.com/blang/semver/v4"
	"github.com/xmidt-org/carousel/pkg/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateSteps(t *testing.T) {
	tests := []struct {
		name          string
		sourceCluster model.ClusterState
		targetCluster model.ClusterState
		options       []StepOptions
		expectedSteps []model.Step
	}{
		{
			name:          "empty_cluster",
			sourceCluster: model.NewClusterState(),
			targetCluster: model.ClusterState{
				model.Blue: model.ClusterGroupState{
					Count:   3,
					Version: semver.Version{},
				},
				model.Green: model.ClusterGroupState{},
			},
			expectedSteps: []model.Step{
				{model.Blue: 0, model.Green: 0},
				{model.Blue: 1, model.Green: 0},
				{model.Blue: 2, model.Green: 0},
				{model.Blue: 3, model.Green: 0},
			},
		},
		{
			name: "clear_cluster",
			sourceCluster: model.ClusterState{
				model.Green: model.ClusterGroupState{
					Count:   3,
					Version: semver.Version{},
				},
				model.Blue: model.ClusterGroupState{}, // TODO test empty color set
			},
			targetCluster: model.NewClusterState(), // TODO:  Test empty struct
			expectedSteps: []model.Step{
				{model.Blue: 0, model.Green: 3},
				{model.Blue: 0, model.Green: 2},
				{model.Blue: 0, model.Green: 1},
				{model.Blue: 0, model.Green: 0},
			},
		},
		{
			name: "batch_by_3",
			sourceCluster: model.ClusterState{
				model.Green: model.ClusterGroupState{
					Count:   3,
					Version: semver.Version{},
				},
				model.Blue: model.ClusterGroupState{}, // TODO test empty color set
			},
			targetCluster: model.ClusterState{
				model.Blue: model.ClusterGroupState{
					Count:   6,
					Version: semver.Version{},
				},
				model.Green: model.ClusterGroupState{},
			},
			options: []StepOptions{
				WithBatchSize(3),
			},
			expectedSteps: []model.Step{
				{model.Blue: 0, model.Green: 3},
				{model.Blue: 3, model.Green: 3},
				{model.Blue: 3, model.Green: 0},
				{model.Blue: 6, model.Green: 0},
			},
		},
		{
			name: "batch_by_3_off_ending",
			sourceCluster: model.ClusterState{
				model.Green: model.ClusterGroupState{
					Count:   3,
					Version: semver.Version{},
				},
				model.Blue: model.ClusterGroupState{},
			},
			targetCluster: model.ClusterState{
				model.Blue: model.ClusterGroupState{
					Count:   7,
					Version: semver.Version{},
				},
				model.Green: model.ClusterGroupState{},
			},
			options: []StepOptions{
				WithBatchSize(3),
			},
			expectedSteps: []model.Step{
				{model.Blue: 0, model.Green: 3},
				{model.Blue: 3, model.Green: 3},
				{model.Blue: 3, model.Green: 0},
				{model.Blue: 6, model.Green: 0},
				{model.Blue: 7, model.Green: 0},
			},
		},
		{
			name: "switch_group_same_count_skip_1",
			sourceCluster: model.ClusterState{
				model.Blue: model.ClusterGroupState{
					Count:   4,
					Version: semver.Version{},
				},
				model.Green: model.ClusterGroupState{},
			},
			targetCluster: model.ClusterState{
				model.Green: model.ClusterGroupState{
					Count:   4,
					Version: semver.Version{},
				},
				model.Blue: model.ClusterGroupState{},
			},
			options: []StepOptions{
				WithSkipFirstN(1),
			},
			expectedSteps: []model.Step{
				{model.Green: 0, model.Blue: 4},
				{model.Green: 2, model.Blue: 4},
				{model.Green: 2, model.Blue: 3},
				{model.Green: 3, model.Blue: 3},
				{model.Green: 3, model.Blue: 2},
				{model.Green: 4, model.Blue: 2},
				{model.Green: 4, model.Blue: 0},
			},
		},
		{
			name: "smaller_target_cluster",
			sourceCluster: model.ClusterState{
				model.Green: model.ClusterGroupState{
					Count:   5,
					Version: semver.Version{},
				},
				model.Blue: model.ClusterGroupState{},
			},
			targetCluster: model.ClusterState{
				model.Blue: model.ClusterGroupState{
					Count:   3,
					Version: semver.Version{},
				},
				model.Green: model.ClusterGroupState{},
			},
			expectedSteps: []model.Step{
				{model.Blue: 0, model.Green: 5},
				{model.Blue: 1, model.Green: 5},
				{model.Blue: 1, model.Green: 4},
				{model.Blue: 2, model.Green: 4},
				{model.Blue: 2, model.Green: 3},
				{model.Blue: 3, model.Green: 3},
				{model.Blue: 3, model.Green: 2},
				{model.Blue: 3, model.Green: 1},
				{model.Blue: 3, model.Green: 0},
			},
		},
		{
			name: "empty_skip_3",
			sourceCluster: model.ClusterState{
				model.Blue: model.ClusterGroupState{
					Count:   4,
					Version: semver.Version{},
				},
				model.Green: model.ClusterGroupState{},
			},
			targetCluster: model.NewClusterState(),
			options: []StepOptions{
				WithSkipFirstN(2),
			},
			expectedSteps: []model.Step{
				{model.Green: 0, model.Blue: 4},
				{model.Green: 0, model.Blue: 3},
				{model.Green: 0, model.Blue: 0},
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

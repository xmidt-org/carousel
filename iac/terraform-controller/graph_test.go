package terraform_controller

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/xmidt-org/carousel/iac"
	"github.com/xmidt-org/carousel/runner"
	"testing"
)

var simpleListRunner = simplerunnable{
	Name: "list builder",
	Data: []byte(goodList),
}

var countListRunner = simplerunnable{
	Name: "list count builder",
	Data: []byte(countGoodList),
}

func TestClusterGraph(t *testing.T) {
	tests := []struct {
		name              string
		searchHost        string
		expectedResources []string
		staterunner       runner.Runnable
		graphrunner       runner.Runnable
		expectedErr       error
	}{
		{
			name:              "empty_output",
			searchHost:        "carousel-demo-ffdbb6.example.com",
			expectedResources: []string{},
			staterunner: simplerunnable{
				Name: "testRunner",
				Data: nil,
			},
			expectedErr: iac.ErrGetClusterFailure,
		},
		{
			name:              "empty_json",
			searchHost:        "carousel-demo-ffdbb6.example.com",
			expectedResources: []string{},
			staterunner: simplerunnable{
				Name: "testRunner",
				Data: []byte(`{}`),
			},
			expectedErr: iac.ErrGetClusterFailure,
		},
		{
			name:              "empty_state",
			searchHost:        "carousel-demo-ffdbb6.example.com",
			expectedResources: []string{},
			staterunner: simplerunnable{
				Name: "testRunner",
				Data: []byte(emptyState),
			},
			graphrunner: simpleListRunner,
			expectedErr: errHostNotInGroup,
		},
		{
			name:              "clean_state/index0",
			searchHost:        "carousel-demo-ffdbb6.example.com",
			expectedResources: []string{"module.green.data.null_data_source.name[0]", "module.green.random_id.ID[0]"},
			staterunner: simplerunnable{
				Name: "testRunner",
				Data: []byte(cleanState),
			},
			graphrunner: simpleListRunner,
		},
		{
			name:              "clean_state/index1",
			searchHost:        "carousel-demo-ea9412.example.com",
			expectedResources: []string{"module.green.data.null_data_source.name[1]", "module.green.random_id.ID[1]"},
			staterunner: simplerunnable{
				Name: "testRunner",
				Data: []byte(cleanState),
			},
			graphrunner: simpleListRunner,
		},
		{
			name:              "count/search",
			searchHost:        "carousel-demo-ea9412.example.com",
			expectedResources: []string{"module.green[1].data.null_data_source.name", "module.green[1].random_id.ID"},
			staterunner: simplerunnable{
				Name: "testRunner",
				Data: []byte(cleanState),
			},
			graphrunner: countListRunner,
		},
		{
			name:              "clean_state/bad_search",
			searchHost:        "carousel.example.com",
			expectedResources: []string{},
			staterunner: simplerunnable{
				Name: "testRunner",
				Data: []byte(cleanState),
			},
			graphrunner: simpleListRunner,
			expectedErr: errHostNotInGroup,
		},
		{
			name:              "clean_state/empty_string",
			searchHost:        "",
			expectedResources: []string{},
			staterunner: simplerunnable{
				Name: "testRunner",
				Data: []byte(cleanState),
			},
			graphrunner: simpleListRunner,
			expectedErr: errEmptyHostName,
		},
		{
			name:              "clean_state_failed_graphbuilder",
			searchHost:        "carousel-demo-ffdbb6.example.com",
			expectedResources: []string{},
			staterunner: simplerunnable{
				Name: "testRunner",
				Data: []byte(cleanState),
			},
			graphrunner: simplerunnable{
				Name: ":( failed runner",
				Data: nil,
			},
			expectedErr: errStateList,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert := assert.New(t)
			stateGetter := &tState{stateRunner: test.staterunner}

			graphGetter := tGraph{
				getter:     stateGetter,
				listRunner: test.graphrunner,
			}
			resource, err := graphGetter.GetResourcesForHost(test.searchHost)
			if test.expectedErr != nil {
				assert.True(errors.Is(err, test.expectedErr))
			} else {
				assert.NoError(err)
			}
			assert.ElementsMatch(test.expectedResources, resource)
		})
	}
}

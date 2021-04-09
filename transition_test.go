package carousel

import (
	"github.com/blang/semver/v4"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestSimpleTransition(t *testing.T) {
	assert := assert.New(t)

	stateGetter := &MockClusterGetter{}
	// initial get and initial apply, which changes nothing
	stateGetter.On("GetCluster").Return(emptyCluster, nil).Twice()

	stateGetter.On("GetCluster").Return(Cluster{
		Green: ClusterGroup{
			Hosts:   []string{"carousel-demo-ffdbb6.example.com"},
			Version: semver.MustParse("0.1.0"),
		},
		Blue: ClusterGroup{},
	}, nil).Once()

	tainter := &MockTainter{}
	r := &MockRunner{}
	r.On("Output").Return([]byte("building step"), nil)
	r.On("String").Return("mock runner")

	tTransitioner := tTransition{
		applyRunBuilder: func(target ClusterState, step Step) Runnable {
			assert.NotEmpty(step)
			assert.NotEmpty(target)
			return r
		},
		validatorResource: func(fqdn string) bool {
			assert.Equal("carousel-demo-ffdbb6.example.com", fqdn)
			return true
		},
		tainter:      tainter,
		CG:           stateGetter,
		getGoalState: DefaultGoalState,
		stepOptions:  nil,
		logger:       log.NewNopLogger(),
	}

	err := tTransitioner.Transition(1, semver.MustParse("0.1.0"))
	assert.NoError(err)

	mock.AssertExpectationsForObjects(t, stateGetter, tainter, r)
}

func TestTransitionWithHostValidation(t *testing.T) {
	assert := assert.New(t)
	badHost := "carousel-demo-ea9412.example.com"

	stateGetter := &MockClusterGetter{}
	// initial get and initial apply, which changes nothing
	stateGetter.On("GetCluster").Return(emptyCluster, nil).Twice()

	stateGetter.On("GetCluster").Return(Cluster{
		Green: ClusterGroup{
			Hosts:   []string{badHost},
			Version: semver.MustParse("0.1.0"),
		},
		Blue: ClusterGroup{},
	}, nil).Once()
	stateGetter.On("GetCluster").Return(Cluster{
		Green: ClusterGroup{
			Hosts:   []string{"carousel-demo-ffdbb6.example.com"},
			Version: semver.MustParse("0.1.0"),
		},
		Blue: ClusterGroup{},
	}, nil).Once()

	tainter := &MockTainter{}
	tainter.On("TaintHost", badHost).Return(nil).Once()

	r := &MockRunner{}
	r.On("Output").Return([]byte("building step"), nil)
	r.On("String").Return("mock runner")

	tTransitioner := tTransition{
		applyRunBuilder: func(target ClusterState, step Step) Runnable {
			assert.NotEmpty(step)
			assert.NotEmpty(target)
			return r
		},
		validatorResource: func(fqdn string) bool {
			return !(fqdn == badHost)
		},
		tainter:      tainter,
		CG:           stateGetter,
		getGoalState: DefaultGoalState,
		stepOptions:  nil,
		logger:       log.NewNopLogger(),
	}

	err := tTransitioner.Transition(1, semver.MustParse("0.1.0"))
	assert.NoError(err)

	mock.AssertExpectationsForObjects(t, stateGetter, tainter, r)
}

func TestDryRunTransition(t *testing.T) {
	assert := assert.New(t)

	stateGetter := &MockClusterGetter{}
	// initial get and initial apply, which changes nothing
	stateGetter.On("GetCluster").Return(emptyCluster, nil).Once()

	tainter := &MockTainter{}

	r := &MockRunner{}
	r.On("String").Return("mock runner").Twice()

	tTransitioner := tTransition{
		applyRunBuilder: func(target ClusterState, step Step) Runnable {
			assert.NotEmpty(step)
			assert.NotEmpty(target)
			return r
		},
		validatorResource: func(fqdn string) bool {
			assert.Fail("validator should not have been called. ")
			return true
		},
		tainter:      tainter,
		CG:           stateGetter,
		getGoalState: DefaultGoalState,
		dryRun:       true,
		logger:       log.NewNopLogger(),
	}

	err := tTransitioner.Transition(1, semver.MustParse("0.1.0"))
	assert.NoError(err)

	mock.AssertExpectationsForObjects(t, stateGetter, r)
}

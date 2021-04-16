package carousel

import (
	"github.com/blang/semver/v4"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/xmidt-org/carousel/pkg/model"
	"testing"
)

type noopUI struct {
}

func (n noopUI) Info(s string) {
}

func (n noopUI) Warn(s string) {
}

var emptyCluster = model.NewCluster()

func TestSimpleTransition(t *testing.T) {
	assert := assert.New(t)

	controller := &MockController{}
	// initial get and initial apply, which changes nothing
	controller.On("GetCluster").Return(emptyCluster, nil).Twice()

	controller.On("GetCluster").Return(model.Cluster{
		model.Green: model.ClusterGroup{
			Hosts:   []string{"carousel-demo-ffdbb6.example.com"},
			Version: semver.MustParse("0.1.0"),
		},
		model.Blue: model.ClusterGroup{},
	}, nil).Once()

	r := &MockRunner{}
	r.On("Output").Return([]byte("building step"), nil)
	r.On("String").Return("mock runner")
	controller.On("CreateApply", mock.Anything, mock.Anything).Return(r)

	carousel := Carousel{
		config: Config{
			DryRun: false,
			Validate: func(fqdn string) bool {
				assert.Equal("carousel-demo-ffdbb6.example.com", fqdn)
				return true
			},
		},
		controller: controller,
		logger:     log.NewNopLogger(),
		ui:         noopUI{},
	}

	err := carousel.Rollout(1, semver.MustParse("0.1.0"))
	assert.NoError(err)

	mock.AssertExpectationsForObjects(t, controller, r)
}

func TestTransitionWithHostValidation(t *testing.T) {
	assert := assert.New(t)
	badHost := "carousel-demo-ea9412.example.com"

	controller := &MockController{}
	// initial get and initial apply, which changes nothing
	controller.On("GetCluster").Return(emptyCluster, nil).Twice()

	controller.On("GetCluster").Return(model.Cluster{
		model.Green: model.ClusterGroup{
			Hosts:   []string{badHost},
			Version: semver.MustParse("0.1.0"),
		},
		model.Blue: model.ClusterGroup{},
	}, nil).Once()
	controller.On("GetCluster").Return(model.Cluster{
		model.Green: model.ClusterGroup{
			Hosts:   []string{"carousel-demo-ffdbb6.example.com"},
			Version: semver.MustParse("0.1.0"),
		},
		model.Blue: model.ClusterGroup{},
	}, nil).Once()

	controller.On("TaintHost", badHost).Return(nil).Once()

	r := &MockRunner{}
	r.On("Output").Return([]byte("building step"), nil)
	r.On("String").Return("mock runner")
	controller.On("CreateApply", mock.Anything, mock.Anything).Return(r)

	carousel := Carousel{
		config: Config{
			DryRun: false,
			Validate: func(fqdn string) bool {
				return !(fqdn == badHost)
			},
		},
		controller: controller,
		logger:     log.NewNopLogger(),
		ui:         noopUI{},
	}

	err := carousel.Rollout(1, semver.MustParse("0.1.0"))
	assert.NoError(err)

	mock.AssertExpectationsForObjects(t, controller, r)
}

func TestDryRunTransition(t *testing.T) {
	assert := assert.New(t)

	controller := &MockController{}
	// initial get and initial apply, which changes nothing
	controller.On("GetCluster").Return(emptyCluster, nil).Once()

	r := &MockRunner{}
	r.On("String").Return("mock runner").Twice()
	controller.On("CreateApply", mock.Anything, mock.Anything).Return(r)

	carousel := Carousel{
		config: Config{
			DryRun: true,
			Validate: func(fqdn string) bool {
				assert.Fail("validator should not have been called. ")
				return true
			},
		},
		controller: controller,
		logger:     log.NewNopLogger(),
		ui:         noopUI{},
	}

	err := carousel.Rollout(1, semver.MustParse("0.1.0"))
	assert.NoError(err)

	mock.AssertExpectationsForObjects(t, controller, r)
}

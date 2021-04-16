package carousel

import (
	"errors"
	"fmt"
	"github.com/blang/semver/v4"
	"github.com/go-kit/kit/log"
	"github.com/xmidt-org/carousel/pkg/controller"
	"github.com/xmidt-org/carousel/pkg/goal"
	"github.com/xmidt-org/carousel/pkg/model"
	"github.com/xmidt-org/carousel/pkg/step"
	"os"
)

type UI interface {
	// Info is used for any messages that might appear on standard
	// output.
	Info(string)
	// Warn is used for any warning messages that might appear on standard
	// error.
	Warn(string)
}

type BasicUI struct{}

func (b BasicUI) Info(s string) {
	fmt.Fprintln(os.Stdout, s)
}

func (b BasicUI) Warn(s string) {
	fmt.Fprintln(os.Stderr, s)
}

type Carousel struct {
	logger     log.Logger
	ui         UI
	controller controller.Controller
	config     Config
}

type Config struct {
	DryRun   bool
	Validate HostValidator
}

// HostValidator is a function that Checks if a Host is bad or good.
type HostValidator func(fqdn string) bool

func AsHostValidator(f func(fqdn string) bool) HostValidator {
	return f
}

func NewCarousel(logger log.Logger, ui UI, controller controller.Controller, config Config) (Carousel, error) {
	if logger == nil {
		logger = log.NewNopLogger()
	}
	if ui == nil {
		ui = BasicUI{}
	}
	if controller == nil {
		return Carousel{}, errors.New("controller can't be empty")
	}
	if config.Validate == nil {
		config.Validate = func(fqdn string) bool { return true }
	}
	return Carousel{
		logger:     logger,
		ui:         ui,
		controller: controller,
		config:     config,
	}, nil
}

func (c Carousel) Rollout(nodeCount int, version semver.Version, stepOptions ...step.StepOptions) error {
	if c.controller == nil {
		return errors.New("controller can't be empty")
	}
	// Get the current cluster.
	cc, err := c.controller.GetCluster()
	if err != nil {
		return fmt.Errorf("%w: %v", controller.ErrGetClusterFailure, err)
	}
	// Determine the goal state.
	goalCluster, err := goal.BuildEndState(cc.AsClusterState(), nodeCount, version)
	if err != nil {
		return fmt.Errorf("%w: %v", controller.ErrGoalStateFailure, err)
	}
	currentGroup, _ := cc.AsClusterState().Group()

	// Build the steps to get to goal
	steps := step.CreateSteps(cc.AsClusterState(), goalCluster, stepOptions...)

	return c.transition(cc, currentGroup, steps, goalCluster)
}

func (c Carousel) Resume(startingColor model.Color, steps []model.Step, goalCluster model.ClusterState) error {
	if c.controller == nil {
		return errors.New("controller can't be empty")
	}
	// Get the current cluster.
	cc, err := c.controller.GetCluster()
	if err != nil {
		return fmt.Errorf("%w: %v", controller.ErrGetClusterFailure, err)
	}
	return c.transition(cc, startingColor, steps, goalCluster)
}

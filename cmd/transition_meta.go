package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/xmidt-org/carousel/pkg/carousel"
	"github.com/xmidt-org/carousel/pkg/controller"
	"github.com/xmidt-org/carousel/pkg/controller/terraform"
	"github.com/xmidt-org/carousel/pkg/model"
	"io/ioutil"
	"os"
	"plugin"
)

type TransitionMeta struct {
	Meta
	jsonOutput bool
	fullOutput bool
	notQuiet   bool
	dryRun     bool
	pluginFile string
	outputFile string
}

// transitionFlagSet adds custom flags that are mostly used by commands
// that are used to run a transition operation like rollout or resume.
func (m *TransitionMeta) transitionFlagSet(n string) *pflag.FlagSet {
	cmdFlags := m.extendedFlagSet(n)
	cmdFlags.BoolVar(&m.jsonOutput, "json", false, "json output")
	cmdFlags.BoolVar(&m.fullOutput, "full", false, "print hostnames")
	cmdFlags.BoolVarP(&m.notQuiet, "quiet", "q", false, "print terraform output")
	cmdFlags.BoolVarP(&m.dryRun, "dry-run", "d", false, "print command to be executed")
	cmdFlags.StringVarP(&m.pluginFile, "plugin", "p", "", "golang plugin file for validating hosts")
	cmdFlags.StringVarP(&m.outputFile, "output", "o", "err.json", "output file for steps upon error")

	return cmdFlags
}

func (m *TransitionMeta) getController() controller.Controller {
	m.LoadConfig()
	transitionConfig := terraform.TerraformTransitionConfig{
		AttachStdOut: !m.notQuiet,
		AttachStdErr: true,
		Args:         m.config.BinaryConfig.Args,
	}

	return terraform.BuildController(m.config.BinaryConfig, transitionConfig)
}

func (m *TransitionMeta) getCarousel() carousel.Carousel {
	validator, err := m.extractValidatorFromPlugin()
	if err != nil {
		m.UI.Error(fmt.Sprintf("Failed to load plugin: %s", err.Error()))
	}

	// no validator, return true for each host
	if validator == nil {
		m.UI.Warn("not checking hosts")
		validator = func(fqdn string) bool { return true }
	}

	carousel, err := carousel.NewCarousel(&UILogger{m.UI}, m.UI, m.getController(), carousel.Config{
		DryRun:   m.dryRun,
		Validate: validator,
	})
	if err != nil {
		m.UI.Error(err.Error())
		os.Exit(1)
	}
	return carousel
}

func (m *TransitionMeta) extractValidatorFromPlugin() (carousel.HostValidator, error) {
	if m.pluginFile == "" {
		return nil, nil
	}
	if p, err := plugin.Open(m.pluginFile); err == nil {
		if f, lookupErr := p.Lookup("CheckHost"); lookupErr == nil {
			if checkHostF, ok := f.(func(string) bool); ok {
				if checkHostF != nil {
					return carousel.AsHostValidator(checkHostF), nil
				} else {
					return nil, fmt.Errorf("CheckHost is nil")
				}
			} else if checkHost, ok := f.(carousel.HostValidator); ok {
				if checkHost != nil {
					return checkHost, nil
				} else {
					return nil, fmt.Errorf("CheckHost is nil")
				}
			} else {
				return nil, fmt.Errorf("plugin file %s func CheckHost is not a carousel.HostValidator", m.pluginFile)
			}
		} else {
			return nil, fmt.Errorf("%w: %s", lookupErr, "CheckHost not defined in plugin file")
		}
	} else {
		return nil, err
	}
}

func (m *TransitionMeta) handleExitError(err error) int {
	if err != nil {
		var stepError model.StepError

		if errors.As(err, &stepError) {
			data, _ := json.MarshalIndent(&stepError, "", " ")
			if writeerr := ioutil.WriteFile(m.outputFile, data, 0644); writeerr != nil {
				m.UI.Error(fmt.Sprintf("failed to write to file %s", m.outputFile))
				m.UI.Info(string(data))
			}
		}
		m.UI.Error(fmt.Sprintf("Failed to transition cluster:\n%v", err))
		return 1
	}
	return 0
}

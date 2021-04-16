package main

import (
	"errors"
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/xmidt-org/carousel/pkg/controller/terraform"
	"github.com/xmidt-org/carousel/pkg/runner"
	"io/ioutil"
)

// Meta based of the terraform Meta struct
// https://github.com/hashicorp/terraform/blob/030632e87ec11b7ce74c7c6065b9779e719488f9/command/meta.go#L38
type Meta struct {
	UI cli.Ui // UI for output
	// When this channel is closed, the command will be canceled.
	ShutdownCh <-chan struct{}
	color      bool
	oldUI      cli.Ui
	file       string

	config *Config
}

// process will process the meta-parameters out of the arguments. This
// will potentially modify the args in-place. It will return the resulting
// slice.
func (m *Meta) process(args []string) []string {
	// We do this so that we retain the ability to technically call
	// process multiple times, even if we have no plans to do so
	if m.oldUI != nil {
		m.UI = m.oldUI
	}

	// Set colorization
	m.color = true
	i := 0 // output index
	for _, v := range args {
		if v == "-no-color" {
			m.color = false
		} else {
			// copy and increment index
			args[i] = v
			i++
		}
	}
	args = args[:i]

	// Set the UI
	m.oldUI = m.UI
	if m.color {
		m.UI = &cli.ConcurrentUi{
			Ui: &cli.ColoredUi{
				OutputColor: cli.UiColorNone,
				InfoColor:   cli.UiColorMagenta,
				ErrorColor:  cli.UiColorRed,
				WarnColor:   cli.UiColorYellow,
				Ui:          m.oldUI,
			},
		}
	}

	return args
}

// defaultFlagSet creates a default flag set for commands.
func (m *Meta) defaultFlagSet(n string) *pflag.FlagSet {
	f := pflag.NewFlagSet(n, pflag.ContinueOnError)
	f.SetOutput(ioutil.Discard)

	// Set the default Usage to empty
	f.Usage = func() {}

	return f
}

func (m *Meta) LoadConfig() Config {
	if m.config == nil {
		v := viper.New()
		if m.file != "" {
			v.SetConfigFile(m.file)
		} else {
			v.SetConfigName(applicationName)
			v.AddConfigPath(fmt.Sprintf("/etc/%s", applicationName))
			v.AddConfigPath(fmt.Sprintf("$HOME/.%s", applicationName))
			v.AddConfigPath(".")
		}
		v.AutomaticEnv() // read in environment variables that match

		if err := v.ReadInConfig(); err != nil {
			// Should we print out the config file?
			m.UI.Error(fmt.Sprintf("Failed to read config file: %v", err))
		}
		config := Config{}
		if err := v.Unmarshal(&config); err != nil {
			m.UI.Error(fmt.Sprintf("Failed to read config: %v", err))
		}
		m.config = &config
	}

	if err := terraform.BuildSelectWorkspaceRunner(m.config.BinaryConfig).SelectWorkspace(m.config.Workspace); err != nil {
		var exitErr runner.ExitError

		if errors.As(err, &exitErr) {
			m.UI.Error(fmt.Sprintf("Failed to select workspace %#v", err))
			m.UI.Output(string(exitErr.CapturedErrorOutput))
		} else {
			m.UI.Error(fmt.Sprintf("Failed to select workspace %#v", err))
		}
	} else {
		if m.config.Workspace != "" {
			m.UI.Warn(fmt.Sprintf("using workspace %s", m.config.Workspace))
		}
	}

	return *m.config
}

// extendedFlagSet adds custom flags that are mostly used by commands
// that are used to run an operation like plan or apply.
func (m *Meta) extendedFlagSet(n string) *pflag.FlagSet {
	f := m.defaultFlagSet(n)

	f.StringVarP(&m.file, "file", "f", "", "the configuration file to use.  Overrides the search path.")

	return f
}

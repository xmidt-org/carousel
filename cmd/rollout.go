package main

import (
	"fmt"
	"github.com/blang/semver/v4"
	"strconv"
	"strings"
)

type RolloutCommand struct {
	TransitionMeta
}

func (c *RolloutCommand) Help() string {
	helpText := `
Usage: %s rollout <server count> <version> [options]

 The terraform .tf files must contain a green module and blue module.
 rollout will transition from one group or module to the other. if both are empty, one will be filled with the given number

Options:

  -json       Output the version information as a JSON object.
`
	return strings.TrimSpace(fmt.Sprintf(helpText, applicationName))
}

func (c *RolloutCommand) Synopsis() string {
	return "transition to a new cluster state"
}

func (c *RolloutCommand) Run(args []string) int {
	args = c.Meta.process(args)
	cmdFlags := c.TransitionMeta.transitionFlagSet("rollout")
	cmdFlags.Usage = func() { c.UI.Error(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if cmdFlags.NArg() != 2 {
		c.UI.Error("only two arguments must be provide")
		c.UI.Error(c.Help())
		return 1
	}

	serverCount, err := strconv.Atoi(cmdFlags.Arg(0))
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to determine number of servers to deploy %v", err))
	}
	version, err := semver.Parse(cmdFlags.Arg(1))
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to determine version of servers to deploy %v", err))
	}

	transitionr := c.TransitionMeta.getTransitionr()
	err = transitionr.Transition(serverCount, version)
	return c.handleExitError(err)
}

package main

import (
	"github.com/xmidt-org/carousel"
	"fmt"
	"strings"
)

type TaintCommand struct {
	Meta
}

func (c *TaintCommand) Help() string {
	helpText := `
Usage: %s taint [options]

  Taint a host in the cluster

`
	return strings.TrimSpace(fmt.Sprintf(helpText, applicationName))
}

func (c *TaintCommand) Synopsis() string {
	return "taint a resource in the current cluster"
}

func (c *TaintCommand) Run(args []string) int {
	var quiet bool

	args = c.Meta.process(args)
	cmdFlags := c.Meta.extendedFlagSet("taint")
	cmdFlags.BoolVarP(&quiet, "quiet", "q", false, "print terraform output")

	cmdFlags.Usage = func() { c.UI.Error(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}
	config := c.Meta.LoadConfig()

	hostCount := cmdFlags.NArg()
	if hostCount == 0 {
		c.UI.Error("at least one host must be provide")
		c.UI.Error(c.Help())
		return 1
	}
	clusterGetter := carousel.BuildStateDeterminer(config.BinaryConfig)
	grapher := carousel.BuildClusterGraphRunner(clusterGetter, config.BinaryConfig)
	tainter := carousel.BuildTaintHostRunner(grapher, config.BinaryConfig)

	for i := 0; i < hostCount; i++ {
		hostname := cmdFlags.Arg(i)
		err := tainter.TaintHost(hostname)
		if err != nil {
			c.UI.Error(fmt.Sprintf("Failed to taint host %s: %v", hostname, err))
			return 1
		}
	}

	return 0
}

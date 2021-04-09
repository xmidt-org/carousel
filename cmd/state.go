package main

import (
	"github.com/xmidt-org/carousel"
	"encoding/json"
	"fmt"
	"strings"
)

type StateCommand struct {
	Meta
}

func (c *StateCommand) Help() string {
	helpText := `
Usage: %s state [options]

  Obtain the current cluster and output in a concise format

Options:

  --json       Output the cluster information as a JSON object.
  --full       Output the hostnames of the cluster.

`
	return strings.TrimSpace(fmt.Sprintf(helpText, applicationName))
}

func (c *StateCommand) Synopsis() string {
	return "Show the current state of the cluster"
}

func (c *StateCommand) Run(args []string) int {
	var jsonOutput bool
	var fullOutput bool

	args = c.Meta.process(args)
	cmdFlags := c.Meta.extendedFlagSet("state")
	cmdFlags.BoolVar(&jsonOutput, "json", false, "json output")
	cmdFlags.BoolVar(&fullOutput, "full", false, "print hostnames")
	cmdFlags.Usage = func() { c.UI.Error(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}
	config := c.Meta.LoadConfig()

	cluster, err := carousel.BuildStateDeterminer(config.BinaryConfig).GetCluster()
	if err != nil {
		c.UI.Error(fmt.Sprintf("Failed to get Cluster state: \n %v", err))
		return 1
	}

	if jsonOutput {
		if fullOutput {
			data, err := json.MarshalIndent(cluster, "", "  ")
			if err != nil {
				c.UI.Error(fmt.Sprintf("\nError marshaling JSON: %s", err))
				return 1
			}
			c.UI.Output(string(data))
		} else {
			data, err := json.MarshalIndent(cluster.AsClusterState(), "", "  ")
			if err != nil {
				c.UI.Error(fmt.Sprintf("\nError marshaling JSON: %s", err))
				return 1
			}
			c.UI.Output(string(data))
		}
	} else {
		if fullOutput {
			for _, groupColor := range carousel.ValidColors {
				c.UI.Output(fmt.Sprintf("%s @ %s", groupColor, cluster[groupColor].Version))
				for _, host := range cluster[groupColor].Hosts {
					c.UI.Output(fmt.Sprintf("\t%s", host))
				}
			}
		} else {
			c.UI.Info(cluster.AsClusterState().String())
		}
	}

	return 0
}
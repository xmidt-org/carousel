package main

import (
	"encoding/json"
	"fmt"
	"github.com/xmidt-org/carousel/model"
	"io/ioutil"
	"strings"
)

type ResumeCommand struct {
	TransitionMeta
}

func (c *ResumeCommand) Help() string {
	helpText := `
Usage: %s resume <step_file> [options]

 The terraform .tf files must contain a green module and blue module.
 resume will resume a failed transition from a

Options:

  -json       Output the version information as a JSON object.
`
	return strings.TrimSpace(fmt.Sprintf(helpText, applicationName))
}

func (c *ResumeCommand) Synopsis() string {
	return "resume transition to a new cluster state"
}

func (c *ResumeCommand) Run(args []string) int {
	args = c.Meta.process(args)
	cmdFlags := c.TransitionMeta.transitionFlagSet("resume")
	cmdFlags.Usage = func() { c.UI.Error(c.Help()) }
	if err := cmdFlags.Parse(args); err != nil {
		return 1
	}

	if cmdFlags.NArg() != 1 {
		c.UI.Error("only one arguments must be provide")
		c.UI.Error(c.Help())
		return 1
	}

	data, err := ioutil.ReadFile(cmdFlags.Arg(0))
	if err != nil {
		c.UI.Error(fmt.Sprintf("failed to read error file %v", err))
	}
	var stepError model.StepError
	err = json.Unmarshal(data, &stepError)
	if err != nil {
		c.UI.Error(fmt.Sprintf("failed to read error file %v", err))
		return 1
	}

	err = c.TransitionMeta.getCarousel().Resume(stepError.StartingColorGroup, stepError.TODO, stepError.GoalClusterState)
	return c.handleExitError(err)
}

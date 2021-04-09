package main

import (
	"fmt"
	"runtime"
	"strings"
)

type VersionCommand struct {
	Meta
}

func (c *VersionCommand) Help() string {
	helpText := `
Usage: %s version [options]

  Displays the version of Terraform and all installed plugins
`
	return strings.TrimSpace(fmt.Sprintf(helpText, applicationName))
}

func (c *VersionCommand) Synopsis() string {
	return fmt.Sprintf("Show the current %s version", applicationName)
}

func (c *VersionCommand) Run(args []string) int {
	args = c.Meta.process(args)
	c.UI.Output(BuildVersion())
	return 0
}

func BuildVersion() string {
	str := ""
	str += fmt.Sprintf("%s %s\n", applicationName, Version)
	str += fmt.Sprintf("  go version: \t%s\n", runtime.Version())
	str += fmt.Sprintf("  built time: \t%s\n", BuildTime)
	str += fmt.Sprintf("  git commit: \t%s\n", GitCommit)
	str += fmt.Sprintf("  os/arch: \t%s/%s\n", runtime.GOOS, runtime.GOARCH)

	return str
}

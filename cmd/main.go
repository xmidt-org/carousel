package main

import (
	"github.com/mitchellh/cli"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	applicationName = "carousel"
)

var (
	GitCommit = "undefined"
	Version   = "undefined"
	BuildTime = "undefined"
)

func main() {
	c := cli.NewCLI(applicationName, Version)
	c.Args = os.Args[1:]

	meta := Meta{
		ShutdownCh: makeShutdownCh(),
		oldUI: &cli.BasicUi{
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
			Reader:      os.Stdin,
		},
	}

	commands := map[string]cli.CommandFactory{
		"taint": func() (cli.Command, error) {
			return &TaintCommand{
				Meta: meta,
			}, nil
		},
		"rollout": func() (cli.Command, error) {
			return &RolloutCommand{
				TransitionMeta{
					Meta: meta,
				},
			}, nil
		},
		"resume": func() (cli.Command, error) {
			return &ResumeCommand{
				TransitionMeta{
					Meta: meta,
				},
			}, nil
		},
		"state": func() (cli.Command, error) {
			return &StateCommand{
				Meta: meta,
			}, nil
		},
		"version": func() (cli.Command, error) {
			return &VersionCommand{
				Meta: meta,
			}, nil
		},
	}
	c.Commands = commands

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}

// makeShutdownCh creates an interrupt listener and returns a channel.
// A message will be sent on the channel for every interrupt received.
func makeShutdownCh() <-chan struct{} {
	resultCh := make(chan struct{})

	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		for {
			<-signalCh
			resultCh <- struct{}{}
		}
	}()

	return resultCh
}

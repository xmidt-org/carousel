package runner

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/xmidt-org/carousel/pkg/model"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

var (
	errStartCMDFailure = errors.New("failed to start cmd runnable")
	errGetStdout       = errors.New("failed to get stdout pipe")
	errGetStderr       = errors.New("failed to get stderr pipe")
)

// Runnable is something tha can run multiple times.
type Runnable interface {
	// Output runs and returns output.
	Output() ([]byte, error)

	// String provides a human readable version for debugging
	String() string
}

type cmdRunner struct {
	commandString string
	cmd           exec.Cmd
	options       Options
}

func (c *cmdRunner) String() string {
	return c.commandString
}

func (c *cmdRunner) Output() ([]byte, error) {
	copyCMD := c.cmd
	var stdOutBuf bytes.Buffer
	var stdErrBuf bytes.Buffer

	var errWriter io.Writer
	var outWriter io.Writer

	if c.options.SuppressErrOutput {
		errWriter = io.MultiWriter(&stdErrBuf)
	} else {
		errWriter = io.MultiWriter(os.Stderr, &stdErrBuf)
	}
	if c.options.ShowOutput {
		outWriter = io.MultiWriter(os.Stdout, &stdOutBuf)
	} else {
		outWriter = io.MultiWriter(&stdOutBuf)
	}
	copyCMD.Stdout = outWriter
	copyCMD.Stderr = errWriter

	if err := copyCMD.Start(); err != nil {
		// bad path, binary not executable, &c
		return nil, err
	}

	// Make sure to forward signals to the subcommand.
	cmdChannel := make(chan error) // used for closing the signals forwarder goroutine
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for {
			select {
			case s := <-signalChannel:
				copyCMD.Process.Signal(s)
			case <-cmdChannel:
				return
			}
		}
	}()
	defer func() {
		signal.Stop(signalChannel)
		close(signalChannel)
	}()

	err := copyCMD.Wait()
	cmdChannel <- err

	if err != nil {
		return stdOutBuf.Bytes(), ExitError{
			CapturedError:       err,
			CapturedErrorOutput: stdErrBuf.Bytes(),
		}
	}

	return stdOutBuf.Bytes(), nil
}

type Options struct {
	// Whether to attach stdin configuration. assume false
	Interactive bool

	// Whether to attach stdout configuration. assume false aka hide output
	ShowOutput bool

	// Whether to attach stdout configuration. assume false aka show err output
	SuppressErrOutput bool
}

func (o Options) WithInteractive(interactive bool) Options {
	o.Interactive = interactive
	return o
}

func (o Options) WithShowOutput(showOutput bool) Options {
	o.ShowOutput = showOutput
	return o
}
func (o Options) WithSuppressErrOutput(suppressErrOutput bool) Options {
	o.SuppressErrOutput = suppressErrOutput
	return o
}

// NewCMDRunner builds an exec.Cmd specific Runnable.
// on each run the exec.Cmd is copied so it can be ran again
func NewCMDRunner(dir, binary string, options Options, args ...string) Runnable {
	if binary == "" {
		// TODO: If we support more than terraform, the config should be validated somewhere.
		binary = "terraform"
	}
	cmd := exec.Command(binary, args...)
	cmd.Env = os.Environ()

	workingDir, _ := os.Getwd()
	if dir != "" {
		if _, err := os.Stat(dir); err == nil {
			workingDir = dir
		}
	}
	cmd.Dir = workingDir

	if options.Interactive {
		cmd.Stdin = os.Stdin
	}

	return &cmdRunner{
		commandString: strings.Join(append([]string{binary}, args...), " "),
		cmd:           *cmd,
		options:       options,
	}
}

// AddEnvironment add environment variables to a given Runnable. This will only do something if the runnable was built
// with NewCMDRunner.
func AddEnvironment(runner Runnable, prefix string, environment []model.ValuePair) Runnable {
	if cRun, ok := runner.(*cmdRunner); ok {
		for _, pair := range environment {
			cRun.commandString = fmt.Sprintf("%s%s=xxxx", prefix, pair.Key) + cRun.commandString
			cRun.cmd.Env = append(cRun.cmd.Env, fmt.Sprintf("%s%s=%s", prefix, pair.Key, pair.Value))
		}
		return cRun
	}
	return runner
}

type ExitError struct {
	CapturedError       error
	CapturedErrorOutput []byte
}

func (e ExitError) Error() string {
	return string(e.CapturedErrorOutput)
}

func (e ExitError) Unwrap() error {
	return e.CapturedError
}

func (e ExitError) GetCode() int {
	var exitErr *exec.ExitError
	if errors.As(e.CapturedError, &exitErr) {
		return exitErr.Sys().(syscall.WaitStatus).ExitStatus()
	}
	return 0
}

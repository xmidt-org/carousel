package runner

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/xmidt-org/carousel/model"
	"io"
	"os"
	"os/exec"
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
	cmd          exec.Cmd
	attachStdOut bool
	attachStdErr bool
}

func (c *cmdRunner) String() string {
	return c.cmd.String()
}

func (c *cmdRunner) Output() ([]byte, error) {
	// todo: add logic for sending to stdout as well.
	copyCMD := c.cmd
	stdout, err := copyCMD.StdoutPipe()
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %v", errGetStdout, err)
	}
	stderr, err := copyCMD.StderrPipe()
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %v", errGetStderr, err)
	}
	rd := bufio.NewReader(stdout)
	ed := bufio.NewReader(stderr)

	err = copyCMD.Start()
	if err != nil {
		return []byte{}, fmt.Errorf("%w: %v", errStartCMDFailure, err)
	}

	stdOutData := make([]byte, 0)
	go func() {
		for {
			str, err := rd.ReadString('\n')
			if err != nil {
				if !errors.Is(err, io.EOF) {
					fmt.Fprint(os.Stderr, "Runnable read stdout error: ", err)
				}
				return
			}
			if c.attachStdOut {
				fmt.Fprint(os.Stdout, str)
			}
			stdOutData = append(stdOutData, []byte(str)...)
		}
	}()

	stdErrData := make([]byte, 0)
	go func() {
		for {
			str, err := ed.ReadString('\n')
			if err != nil {
				if !errors.Is(err, io.EOF) {
					fmt.Fprintln(os.Stderr, "Runnable read stderr error: ", err)
				}
				return
			}
			if c.attachStdErr {
				fmt.Fprint(os.Stderr, str)
			}

			stdErrData = append(stdErrData, []byte(str)...)
		}
	}()
	err = copyCMD.Wait()
	if err != nil {
		return stdOutData, ExitError{
			CapturedError:       err,
			CapturedErrorOutput: stdErrData,
		}
	}
	return stdOutData, nil
}

// NewCMDRunner builds an exec.Cmd specific Runnable.
// on each run the exec.Cmd is copied so it can be ran again
func NewCMDRunner(dir, binary string, attachStdin bool, attachStdOut bool, attachStdErr bool, args ...string) Runnable {
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

	if attachStdin {
		cmd.Stdin = os.Stdin
	}

	return &cmdRunner{
		cmd:          *cmd,
		attachStdOut: attachStdOut,
		attachStdErr: attachStdErr,
	}
}

// AddEnvironment add environment variables to a given Runnable. This will only do something if the runnable was built
// with NewCMDRunner.
func AddEnvironment(runner Runnable, prefix string, environment []model.ValuePair) Runnable {
	if cRun, ok := runner.(*cmdRunner); ok {
		for _, pair := range environment {
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

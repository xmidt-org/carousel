package model

import (
	"strings"
)

// MultiError is an interface that provides a list of errors.
type MultiError interface {
	Errors() []error
}

// Errors is a Multierror that also acts as an error, so that a log-friendly
// string can be returned but each error in the list can also be accessed.
type Errors []error

// Error concatenates the list of error strings to provide a single string
// that can be used to represent the errors that occurred.
func (e Errors) Error() string {
	var output strings.Builder
	output.Write([]byte("multiple errors: ["))
	for i, msg := range e {
		if i > 0 {
			output.WriteRune(',')
			output.WriteRune(' ')
		}

		output.WriteString(msg.Error())
	}
	output.WriteRune(']')

	return output.String()
}

// Errors returns the list of errors.
func (e Errors) Errors() []error {
	return e
}

type StepError struct {
	Cause              error        `json:"-"`
	TODO               []Step       `json:"todo"`
	OriginalCluster    Cluster      `json:"original_cluster"`
	StartingColorGroup Color        `json:"starting_group"`
	GoalClusterState   ClusterState `json:"goal_state"`
}

func (e StepError) Error() string {
	return e.Cause.Error()
}
func (e *StepError) Unwrap() error {
	return e.Cause
}

type RunnableError struct {
	Output    []byte
	ResultErr error
}

func (e RunnableError) Error() string {
	return e.ResultErr.Error()
}

func (e *RunnableError) Unwrap() error {
	return e.ResultErr
}

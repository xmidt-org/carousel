package carousel

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"os/exec"
	"testing"
)

func TestTainter(t *testing.T) {
	assert := assert.New(t)
	fakeClusterGraph := &MockClusterGraph{}
	fakeClusterGraph.On("GetResourcesForHost", "").Return([]string{}, errors.New("host can't be empty")).Once()
	fakeClusterGraph.On("GetResourcesForHost", mock.Anything).Return([]string{"asset1", "id2"}, nil).Times(3)
	mockRunner := &MockRunner{}
	// TaintResources all request succeeded
	mockRunner.On("Output").Return(nil, nil).Times(2)

	// TaintResources failed request
	mockRunner.On("Output").Return(nil, &exec.ExitError{
		ProcessState: &os.ProcessState{},
		Stderr:       []byte("asset1 cannot be tainted"),
	}).Once()
	// Second Taint Host, 1 resources succeed, the other failed
	mockRunner.On("Output").Return(nil, nil).Once()
	mockRunner.On("Output").Return(nil, ExitError{
		CapturedError:       errors.New("exit status 1"),
		CapturedErrorOutput: []byte("asset1 cannot be tainted"),
	}).Once()
	// Lasts TaintHost, failed to run binary
	mockRunner.On("String").Return("mockRunner")
	mockRunner.On("Output").Return(nil, &exec.ExitError{
		ProcessState: &os.ProcessState{},
		Stderr:       []byte("binary not found"),
	}).Once()

	tainter := tTaint{
		graphCluster: fakeClusterGraph,
		taintRunnerBuilder: func(key string) Runnable {
			if key == "" {
				assert.Fail("key should not be empty")
			}
			return mockRunner
		},
	}
	err := tainter.TaintHost("")
	assert.Error(err)

	err = tainter.TaintHost("carousel-demo-ffdbb6.example.com")
	assert.NoError(err)
	err = tainter.TaintHost("carousel-demo-ffdbb6.example.com")
	assert.NoError(err)
	err = tainter.TaintHost("carousel-demo-ffdbb6.example.com")
	assert.Error(err)

	mock.AssertExpectationsForObjects(t, fakeClusterGraph, mockRunner)
}

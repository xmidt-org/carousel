package carousel

import (
	"github.com/stretchr/testify/mock"
	"github.com/xmidt-org/carousel/model"
	"github.com/xmidt-org/carousel/runner"
)

type MockController struct {
	mock.Mock
}

func (m *MockController) SelectWorkspace(workspace string) error {
	args := m.Called(workspace)
	return args.Error(0)
}

func (m *MockController) CreateApply(target model.ClusterState, step model.Step) runner.Runnable {
	args := m.Called(target, step)
	return args.Get(0).(runner.Runnable)
}

func (m *MockController) GetCluster() (model.Cluster, error) {
	args := m.Called()
	return args.Get(0).(model.Cluster), args.Error(1)
}

func (m *MockController) TaintResources(resources []string) error {
	args := m.Called(resources)
	return args.Error(0)
}

func (m *MockController) TaintHost(hostname string) error {
	args := m.Called(hostname)
	return args.Error(0)
}

type MockRunner struct {
	mock.Mock
}

func (m *MockRunner) Output() ([]byte, error) {
	args := m.Called()
	if data := args.Get(0); data != nil {
		return args.Get(0).([]byte), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRunner) String() string {
	args := m.Called()
	return args.String(0)
}

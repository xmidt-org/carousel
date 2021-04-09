package carousel

import "github.com/stretchr/testify/mock"

type MockClusterGraph struct {
	mock.Mock
}

func (m *MockClusterGraph) GetResourcesForHost(hostname string) ([]string, error) {
	args := m.Called(hostname)
	if resources := args.Get(0); resources != nil {
		return args.Get(0).([]string), args.Error(1)
	}
	return nil, args.Error(1)
}

type MockClusterGetter struct {
	mock.Mock
}

func (m *MockClusterGetter) GetCluster() (Cluster, error) {
	args := m.Called()
	return args.Get(0).(Cluster), args.Error(1)
}

type MockTainter struct {
	mock.Mock
}

func (m *MockTainter) TaintResources(resources []string) error {
	args := m.Called(resources)
	return args.Error(0)
}

func (m *MockTainter) TaintHost(hostname string) error {
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

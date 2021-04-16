package terraform

import (
	"errors"
	"fmt"
	"github.com/xmidt-org/carousel/pkg/controller"
	"github.com/xmidt-org/carousel/pkg/model"
	"github.com/xmidt-org/carousel/pkg/runner"
	"strings"
)

type tSelectWorkspace struct {
	initRunner            runner.Runnable
	showRunner            runner.Runnable
	listRunner            runner.Runnable
	selectWorkspaceRunner func(workspace string) runner.Runnable
	newWorkspaceRunner    func(workspace string) runner.Runnable
}

var (
	errInitWorkspaceFailure   = errors.New("failed to initialize workspace")
	errShowWorkspaceFailure   = errors.New("failed to show workspace")
	errListWorkspaceFailure   = errors.New("failed to list workspace")
	errSelectWorkspaceFailure = errors.New("failed to select workspace")
	errCreateWorkspaceFailure = errors.New("failed to create workspace")
)

func (t *tSelectWorkspace) SelectWorkspace(workspace string) error {
	_, err := t.initRunner.Output()
	if err != nil {
		return fmt.Errorf("%w: %v", errInitWorkspaceFailure, err)
	}
	if workspace == "" {
		// if workspace is empty use what is currently selected
		return nil
	}

	data, err := t.showRunner.Output()
	if err != nil {
		return fmt.Errorf("%w: %v", errShowWorkspaceFailure, err)
	}
	// already in current workspace
	if strings.TrimSpace(string(data)) == workspace {
		return nil
	}

	// does the workspace exist?
	data, err = t.listRunner.Output()
	if err != nil {
		return fmt.Errorf("%w: %v", errListWorkspaceFailure, err)
	}

	// select workspace
	if strings.Contains(string(data), workspace) {
		// switch workspace
		_, err = t.selectWorkspaceRunner(workspace).Output()
		if err != nil {
			return fmt.Errorf("%v: %w for %s", errSelectWorkspaceFailure, err, workspace)
		}
		return nil
	}

	// create workspace
	_, err = t.newWorkspaceRunner(workspace).Output()
	if err != nil {
		return fmt.Errorf("%v: %w for %s", errCreateWorkspaceFailure, err, workspace)
	}
	return nil
}

// BuildStateDeterminer builds a terraform specific WorkspaceSelecter.
func BuildSelectWorkspaceRunner(config model.BinaryConfig) controller.WorkspaceSelecter {
	return &tSelectWorkspace{
		initRunner: runner.NewCMDRunner(config.WorkingDirectory, config.Binary, false, false, true, "init"),
		showRunner: runner.NewCMDRunner(config.WorkingDirectory, config.Binary, false, false, false, "workspace", "show"),
		listRunner: runner.NewCMDRunner(config.WorkingDirectory, config.Binary, false, false, false, "workspace", "list"),
		selectWorkspaceRunner: func(workspace string) runner.Runnable {
			return runner.NewCMDRunner(config.WorkingDirectory, config.Binary, false, false, true, "workspace", "select", workspace)
		},
		newWorkspaceRunner: func(workspace string) runner.Runnable {
			return runner.NewCMDRunner(config.WorkingDirectory, config.Binary, false, false, true, "workspace", "new", workspace)
		},
	}
}

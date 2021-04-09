package carousel

import (
	"errors"
	"fmt"
	"strings"
)

type tSelectWorkspace struct {
	initRunner            Runnable
	showRunner            Runnable
	listRunner            Runnable
	selectWorkspaceRunner func(workspace string) Runnable
	newWorkspaceRunner    func(workspace string) Runnable
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

// BuildStateDeterminer builds a terraform specific SelectWorkspace.
func BuildSelectWorkspaceRunner(config BinaryConfig) SelectWorkspace {
	return &tSelectWorkspace{
		initRunner: NewCMDRunner(config.WorkingDirectory, config.Binary, false, false, true, "init"),
		showRunner: NewCMDRunner(config.WorkingDirectory, config.Binary, false, false, false, "workspace", "show"),
		listRunner: NewCMDRunner(config.WorkingDirectory, config.Binary, false, false, false, "workspace", "list"),
		selectWorkspaceRunner: func(workspace string) Runnable {
			return NewCMDRunner(config.WorkingDirectory, config.Binary, false, false, true, "workspace", "select", workspace)
		},
		newWorkspaceRunner: func(workspace string) Runnable {
			return NewCMDRunner(config.WorkingDirectory, config.Binary, false, false, true, "workspace", "new", workspace)
		},
	}
}

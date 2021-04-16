package terraform

import (
	"github.com/xmidt-org/carousel/pkg/controller"
	"github.com/xmidt-org/carousel/pkg/model"
)

func BuildController(config model.BinaryConfig, transitionConfig TerraformTransitionConfig) controller.Controller {
	clusterGetter := BuildStateDeterminer(config)
	grapher := BuildClusterGraphRunner(clusterGetter, config)
	tainter := BuildTaintHostRunner(grapher, config)

	return struct {
		controller.WorkspaceSelecter
		controller.ClusterGetter
		controller.Tainter
		controller.ApplyBuilder
	}{
		WorkspaceSelecter: BuildSelectWorkspaceRunner(config),
		ClusterGetter:     clusterGetter,
		Tainter:           tainter,
		ApplyBuilder:      BuildTransitioner(config, transitionConfig),
	}
}

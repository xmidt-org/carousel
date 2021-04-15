package terraform_controller

import (
	"github.com/xmidt-org/carousel/iac"
	"github.com/xmidt-org/carousel/model"
)

func BuildController(config model.BinaryConfig, transitionConfig TerraformTransitionConfig) iac.Controller {
	clusterGetter := BuildStateDeterminer(config)
	grapher := BuildClusterGraphRunner(clusterGetter, config)
	tainter := BuildTaintHostRunner(grapher, config)

	return struct {
		iac.WorkspaceSelecter
		iac.ClusterGetter
		iac.Tainter
		iac.ApplyBuilder
	}{
		WorkspaceSelecter: BuildSelectWorkspaceRunner(config),
		ClusterGetter:     clusterGetter,
		Tainter:           tainter,
		ApplyBuilder:      BuildTransitioner(config, transitionConfig),
	}
}

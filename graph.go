package carousel

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	errEmptyHostName     = errors.New("hostname can not be empty")
	errStateList         = errors.New("failed to list resources")
	errHostNotInGroup    = errors.New("host not part of group")
	errGetClusterFailure = errors.New("failed to get cluster state")
)

type tGraph struct {
	getter     ClusterGetter
	listRunner Runnable
}

func (t *tGraph) GetResourcesForHost(hostname string) ([]string, error) {
	if hostname == "" {
		return []string{}, errEmptyHostName
	}
	c, err := t.getter.GetCluster()
	if err != nil {
		return []string{}, fmt.Errorf("%w: %v", errGetClusterFailure, err)
	}
	group := Unknown
	index := 0

LOOP:
	for color, clusterGroup := range c {
		for i, host := range clusterGroup.Hosts {
			if host == strings.TrimSpace(hostname) {
				group = color
				index = i
				break LOOP
			}
		}
	}
	if group == Unknown {
		return []string{}, fmt.Errorf("%w: %v", errHostNotInGroup, hostname)
	}

	listBytes, err := t.listRunner.Output()
	if err != nil {
		return []string{}, fmt.Errorf("%w: %v", errStateList, err)
	}
	clusterResources := strings.Split(string(listBytes), "\n")

	resourceRegex, err := regexp.Compile(".*?\\[" + strconv.Itoa(index) + "\\].*")
	if err != nil {
		return []string{}, err
	}

	resources := make([]string, 0)
	for _, resource := range clusterResources {
		if resourceRegex.MatchString(resource) {
			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// BuildStateDeterminer builds a terraform specific ClusterGraph.
func BuildClusterGraphRunner(getter ClusterGetter, config BinaryConfig) ClusterGraph {
	return &tGraph{
		getter:     getter,
		listRunner: NewCMDRunner(config.WorkingDirectory, config.Binary, false, false, true, "state", "list"),
	}
}

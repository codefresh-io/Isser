/*
Copyright 2019 The Codefresh Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package plugins

import (
	"fmt"
	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"
)

// k8sAgentPlugin installs assets on Kubernetes Dind runtimectl Env
type k8sAgentPlugin struct {
	logger logger.Logger
}

const (
	k8sAgentFilesPattern = ".*.k8sagent.yaml"
)

// Install k8sAgent agent
func (u *k8sAgentPlugin) Install(opt *InstallOptions, v Values) (Values, error) {

	cs, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil, err
	}
	err = opt.KubeBuilder.EnsureNamespaceExists(cs)
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot ensure namespace exists: %v", err))
		return nil, err
	}
	return v, install(&installOptions{
		logger:         u.logger,
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      opt.ClusterNamespace,
		matchPattern:   k8sAgentFilesPattern,
		dryRun:         opt.DryRun,
		operatorType:   K8sAgentPluginType,
	})
}

func (u *k8sAgentPlugin) Status(statusOpt *StatusOptions, v Values) ([][]string, error) {
	return nil, nil
}

func (u *k8sAgentPlugin) Delete(deleteOpt *DeleteOptions, v Values) error {
	return nil
}

func (u *k8sAgentPlugin) Upgrade(opt *UpgradeOptions, v Values) (Values, error) {
	return nil, nil
}
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
	"encoding/base64"
	"fmt"

	"github.com/codefresh-io/venona/venonactl/pkg/codefresh"
	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"
)

// runtimeEnvironmentPlugin installs assets on Kubernetes Dind runtimectl Env
type runtimeEnvironmentPlugin struct {
	logger logger.Logger
}

const (
	runtimeEnvironmentFilesPattern = ".*.re.yaml"
)

// Install runtimectl environment
func (u *runtimeEnvironmentPlugin) Install(opt *InstallOptions, v Values) (Values, error) {
	cs, err := opt.KubeBuilder.BuildClient()
	if err != nil {
		return nil, fmt.Errorf("Cannot create kubernetes clientset: %v ", err)
	}

	cfOpt := &codefresh.APIOptions{
		Logger:                u.logger,
		CodefreshHost:         opt.CodefreshHost,
		CodefreshToken:        opt.CodefreshToken,
		ClusterName:           opt.ClusterName,
		RegisterWithAgent:     opt.RegisterWithAgent,
		ClusterNamespace:      opt.ClusterNamespace,
		MarkAsDefault:         opt.MarkAsDefault,
		StorageClass:          opt.StorageClass,
		IsDefaultStorageClass: opt.IsDefaultStorageClass,
		KubernetesRunnerType:  opt.KubernetesRunnerType,
	}

	cf := codefresh.NewCodefreshAPI(cfOpt)
	cert, err := cf.Sign()
	if err != nil {
		return nil, err
	}
	v["ServerCert"] = map[string]string{
		"Cert": base64.StdEncoding.EncodeToString([]byte(cert.Cert)),
		"Key":  base64.StdEncoding.EncodeToString([]byte(cert.Key)),
		"Ca":   base64.StdEncoding.EncodeToString([]byte(cert.Ca)),
	}

	if err := cf.Validate(); err != nil {
		return nil, err
	}

	err = install(&installOptions{
		logger:         u.logger,
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      opt.ClusterNamespace,
		matchPattern:   runtimeEnvironmentFilesPattern,
		operatorType:   RuntimeEnvironmentPluginType,
		dryRun:         opt.DryRun,
	})
	if err != nil {
		return nil, err
	}

	re, err := cf.Register()
	if err != nil {
		return nil, err
	}
	v["RuntimeEnvironment"] = re.Metadata.Name

	return v, nil
}

func (u *runtimeEnvironmentPlugin) Status(statusOpt *StatusOptions, v Values) ([][]string, error) {
	cs, err := statusOpt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil, err
	}
	opt := &statusOptions{
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      statusOpt.ClusterNamespace,
		matchPattern:   runtimeEnvironmentFilesPattern,
		operatorType:   RuntimeEnvironmentPluginType,
		logger:         u.logger,
	}
	return status(opt)
}

func (u *runtimeEnvironmentPlugin) Delete(deleteOpt *DeleteOptions, v Values) error {
	cs, err := deleteOpt.KubeBuilder.BuildClient()
	if err != nil {
		u.logger.Error(fmt.Sprintf("Cannot create kubernetes clientset: %v ", err))
		return nil
	}
	opt := &deleteOptions{
		templates:      templates.TemplatesMap(),
		templateValues: v,
		kubeClientSet:  cs,
		namespace:      deleteOpt.ClusterNamespace,
		matchPattern:   runtimeEnvironmentFilesPattern,
		operatorType:   RuntimeEnvironmentPluginType,
		logger:         u.logger,
	}
	return delete(opt)
}

func (u *runtimeEnvironmentPlugin) Upgrade(_ *UpgradeOptions, v Values) (Values, error) {
	return v, nil
}

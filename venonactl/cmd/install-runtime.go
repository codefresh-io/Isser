package cmd

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

import (
	"fmt"

	"github.com/codefresh-io/venona/venonactl/pkg/plugins"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var installRuntimeCmdOptions struct {
	agentID string
	agentToken string
	dryRun bool
	kube   struct {
		namespace string
		inCluster bool
		context   string
	}
	storageClass           string
	runtimeEnvironmentName string
	kubernetesRunnerType   bool
	
}

var installRuntimeCmd = &cobra.Command{
	Use:   "runtime",
	Short: "Install Codefresh's runtime",
	Run: func(cmd *cobra.Command, args []string) {

		s := store.GetStore()
		lgr := createLogger("Install-runtime", verbose)
		buildBasicStore(lgr)
		extendStoreWithAgentAPI(lgr, installRuntimeCmdOptions.agentToken, installRuntimeCmdOptions.agentID)
		extendStoreWithKubeClient(lgr)

		if installRuntimeCmdOptions.runtimeEnvironmentName == "" {
			dieOnError(fmt.Errorf("Codefresh envrionment name is required"))
		}
		if cfAPIHost == "" {
			cfAPIHost = "https://g.codefresh.io"
		}
		// This is temporarily and used for signing
		s.CodefreshAPI = &store.CodefreshAPI{
			Host: cfAPIHost,
		}

		builder := plugins.NewBuilder(lgr)
		isDefault := isUsingDefaultStorageClass(installRuntimeCmdOptions.storageClass)

		builderInstallOpt := &plugins.InstallOptions{
			StorageClass:          installRuntimeCmdOptions.storageClass,
			IsDefaultStorageClass: isDefault,
			DryRun:                installRuntimeCmdOptions.dryRun,
			KubernetesRunnerType:  installRuntimeCmdOptions.kubernetesRunnerType,
			CodefreshHost:         cfAPIHost,
			CodefreshToken:        installRuntimeCmdOptions.agentToken,
			RuntimeEnvironment:    installRuntimeCmdOptions.runtimeEnvironmentName,
			ClusterNamespace:      installRuntimeCmdOptions.kube.namespace,
		}

		if installRuntimeCmdOptions.kubernetesRunnerType {
			builder.Add(plugins.EnginePluginType)
		}

		if isDefault {
			builderInstallOpt.StorageClass = plugins.DefaultStorageClassNamePrefix
		}

		fillKubernetesAPI(lgr, installRuntimeCmdOptions.kube.context, installRuntimeCmdOptions.kube.namespace, installRuntimeCmdOptions.kube.inCluster)

		if installRuntimeCmdOptions.dryRun {
			s.DryRun = installRuntimeCmdOptions.dryRun
			lgr.Info("Running in dry-run mode")
		}

		// s.ClusterInCodefresh = installRuntimeCmdOptions.clusterNameInCodefresh

		builder.Add(plugins.RuntimeEnvironmentPluginType)

		if isDefault {
			builder.Add(plugins.VolumeProvisionerPluginType)
		} else {
			lgr.Info("Custom StorageClass is set, skipping installation of default volume provisioner")
		}

		builderInstallOpt.KubeBuilder = getKubeClientBuilder(builderInstallOpt.ClusterName, s.KubernetesAPI.Namespace, s.KubernetesAPI.ConfigPath, s.KubernetesAPI.InCluster)
		var err error
		values := s.BuildValues()
		for _, p := range builder.Get() {
			values, err = p.Install(builderInstallOpt, values)
			if err != nil {
				dieOnError(err)
			}
		}
		lgr.Info("Runtime installation completed Successfully")

	},
}

func init() {
	installCommand.AddCommand(installRuntimeCmd)

	viper.BindEnv("kube-namespace", "KUBE_NAMESPACE")
	viper.BindEnv("kube-context", "KUBE_CONTEXT")

	installRuntimeCmd.Flags().StringVar(&installRuntimeCmdOptions.agentToken, "agentToken", "", "Agent token created by codefresh")
	installRuntimeCmd.Flags().StringVar(&installRuntimeCmdOptions.agentID, "agentId", "", "Agent id created by codefresh")
	installRuntimeCmd.Flags().StringVar(&installRuntimeCmdOptions.runtimeEnvironmentName, "runtimeName", viper.GetString("runtimeName"), "Name of the runtime as in codefresh")
	installRuntimeCmd.Flags().StringVar(&installRuntimeCmdOptions.kube.namespace, "kube-namespace", viper.GetString("kube-namespace"), "Name of the namespace on which venona should be installed [$KUBE_NAMESPACE]")
	installRuntimeCmd.Flags().StringVar(&installRuntimeCmdOptions.kube.context, "kube-context-name", viper.GetString("kube-context"), "Name of the kubernetes context on which venona should be installed (default is current-context) [$KUBE_CONTEXT]")
	installRuntimeCmd.Flags().StringVar(&installRuntimeCmdOptions.storageClass, "storage-class", "", "Set a name of your custom storage class, note: this will not install volume provisioning components")

	installRuntimeCmd.Flags().BoolVar(&installRuntimeCmdOptions.kube.inCluster, "in-cluster", false, "Set flag if venona is been installed from inside a cluster")
	installRuntimeCmd.Flags().BoolVar(&installRuntimeCmdOptions.dryRun, "dry-run", false, "Set to true to simulate installation")
	installRuntimeCmd.Flags().BoolVar(&installRuntimeCmdOptions.kubernetesRunnerType, "kubernetes-runner-type", false, "Set the runner type to kubernetes (alpha feature)")

}

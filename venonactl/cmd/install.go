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

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/codefresh-io/venona/venonactl/pkg/store"

	"github.com/codefresh-io/venona/venonactl/internal"

	"github.com/codefresh-io/venona/venonactl/pkg/codefresh"
	runtimectl "github.com/codefresh-io/venona/venonactl/pkg/operators"
	"github.com/spf13/cobra"
)

var (
	clusterName                   string
	dryRun                        bool
	skipRuntimeInstallation       bool
	installOnlyRuntimeEnvironment bool
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Codefresh's runtime-environment",
	Run: func(cmd *cobra.Command, args []string) {
		s := store.GetStore()

		kubeContextName := cmd.Flag("kube-context-name").Value.String()
		kubeNamespace := cmd.Flag("kube-namespace").Value.String()

		if kubeContextName == "" {
			config := clientcmd.GetConfigFromFileOrDie(s.KubernetesAPI.ConfigPath)
			kubeContextName = config.CurrentContext
			logrus.WithFields(logrus.Fields{
				"Kube-Context-Name": kubeContextName,
			}).Debug("Kube Context is not set, using current context")
		}
		s.KubernetesAPI.ContextName = kubeContextName
		s.KubernetesAPI.Namespace = kubeNamespace

		version := cmd.Flag("venona-version").Value.String()
		if dryRun == true {
			s.DryRun = dryRun
			logrus.Info("Running in dry-run mode")
		}
		if version != "" {
			logrus.WithFields(logrus.Fields{
				"venona-Version": version,
			}).Info("venona version set by user")
			s.Image.Tag = version
			s.Version.Latest.Version = version
		}
		s.ClusterInCodefresh = clusterName
		if installOnlyRuntimeEnvironment == true && skipRuntimeInstallation == true {
			internal.DieOnError(fmt.Errorf("Cannot use both flags skip-runtime-installation and only-runtime-environment"))
		}
		if installOnlyRuntimeEnvironment == true {
			installRuntimeEnvironment()
			return
		} else if skipRuntimeInstallation == true {
			runtimeEnvironmentName := cmd.Flag("runtime-environment").Value.String()
			if runtimeEnvironmentName == "" {
				internal.DieOnError(fmt.Errorf("runtime-environment flag is required when using flag skip-runtime-installation"))
			}
			s.RuntimeEnvironment = runtimeEnvironmentName
			logrus.Info("Skipping installation of runtime environment, installing venona only")
			installvenona()
		} else {
			installRuntimeEnvironment()
			installvenona()
		}
		logrus.Info("Installation completed Successfully\n")
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().StringVar(&clusterName, "cluster-name", "", "cluster name (if not passed runtime-environment will be created cluster-less)")
	installCmd.Flags().String("venona-version", "", "Version of venona to install (default is the latest)")
	installCmd.Flags().BoolVar(&skipRuntimeInstallation, "skip-runtime-installation", false, "Set flag if you already have a configured runtime-environment, add --runtime-environment flag with name")
	installCmd.Flags().String("runtime-environment", "", "if --skip-runtime-installation set, will try to configure venona on current runtime-environment")
	installCmd.Flags().BoolVar(&installOnlyRuntimeEnvironment, "only-runtime-environment", false, "Set to true to onlky configure namespace as runtime-environment for Codefresh")
	installCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Set to true to simulate installation")
	installCmd.Flags().String("kube-namespace", "default", "Name of the namespace on which venona should be installed")
	installCmd.Flags().String("kube-context-name", "", "Name of the kubernetes context on which venona should be installed (default is current-context)")
}

func installRuntimeEnvironment() {
	cfAPI := codefresh.New()
	err := cfAPI.Validate()
	internal.DieOnError(err)

	err = cfAPI.Sign()
	internal.DieOnError(err)

	err = runtimectl.GetOperator(runtimectl.RuntimeEnvironmentOperatorType).Install()
	internal.DieOnError(err)

	err = cfAPI.Register()
	internal.DieOnError(err)

}

func installvenona() {
	err := runtimectl.GetOperator(runtimectl.VenonaOperatorType).Install()
	internal.DieOnError(err)
}

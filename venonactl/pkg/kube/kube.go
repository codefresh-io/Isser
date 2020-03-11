package kube

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1Core "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	Kube interface {
		BuildClient() (*kubernetes.Clientset, error)
		BuildConfig() (clientcmd.ClientConfig)
		EnsureNamespaceExists(cs *kubernetes.Clientset) (error)
	}

	kube struct {
		contextName      string
		namespace        string
		pathToKubeConfig string
		inCluster        bool
	}

	Options struct {
		ContextName      string
		Namespace        string
		PathToKubeConfig string
		InCluster        bool
	}
)

func New(o *Options) Kube {
	return &kube{
		contextName:      o.ContextName,
		namespace:        o.Namespace,
		pathToKubeConfig: o.PathToKubeConfig,
		inCluster:        o.InCluster,
	}
}

func (k *kube) BuildClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	if k.inCluster {
		config, err = rest.InClusterConfig()
	} else {
		config, err = k.BuildConfig().ClientConfig()
	}
	if err != nil {
		return nil, err
	}
	cs, err := kubernetes.NewForConfig(config)
	return cs, nil
}

func (k *kube) EnsureNamespaceExists(cs *kubernetes.Clientset) (error) {
	_, err := cs.CoreV1().Namespaces().Get(k.namespace, v1.GetOptions{})
	if err != nil  {
		nsSpec := &v1Core.Namespace{ObjectMeta: metav1.ObjectMeta {Name: k.namespace}};
		_, err := cs.CoreV1().Namespaces().Create(nsSpec)
		if err != nil {
			return  err
		} 
	}
	return nil
}

func (k *kube) BuildConfig() (clientcmd.ClientConfig) {
	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: k.pathToKubeConfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: k.contextName,
			Context: clientcmdapi.Context{
				Namespace: k.namespace,
			},
		})
		return config

}

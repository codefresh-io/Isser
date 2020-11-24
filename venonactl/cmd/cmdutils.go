package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"encoding/json"

	"github.com/briandowns/spinner"
	"github.com/codefresh-io/go-sdk/pkg/codefresh"
	sdkUtils "github.com/codefresh-io/go-sdk/pkg/utils"
	"github.com/codefresh-io/venona/venonactl/pkg/certs"
	"github.com/codefresh-io/venona/venonactl/pkg/kube"
	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	"github.com/codefresh-io/venona/venonactl/pkg/plugins"
	"github.com/codefresh-io/venona/venonactl/pkg/store"
	"github.com/olekukonko/tablewriter"
	k8sApi "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/stretchr/objx"
	cliValues "helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"

	verbose      bool
	logFormatter string

	configPath string
	cfAPIHost  string
	cfAPIToken string
	cfContext  string

	kubeConfigPath string
)

func buildBasicStore(logger logger.Logger) {
	s := store.GetStore()
	s.Version = &store.Version{
		Current: &store.CurrentVersion{
			Version: version,
			Commit:  commit,
			Date:    date,
		},
	}

	s.Image = &store.Image{
		Name: "quay.io/codefresh/venona",
	}

	s.Mode = store.ModeInCluster

	s.ServerCert = &certs.ServerCert{}

	s.AppName = store.ApplicationName
	s.Image.Tag = s.Version.Current.Version
}

func extendStoreWithCodefershClient(logger logger.Logger) error {
	s := store.GetStore()
	if configPath == "" {
		currentUser, err := user.Current()
		if err != nil {
			return err
		}

		configPath = filepath.Join(currentUser.HomeDir, ".cfconfig")
		logger.Debug(fmt.Sprint("cfconfig path not set, using: ", configPath))
	}

	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf(".cfconfig file not found")
	}

	if cfAPIHost == "" && cfAPIToken == "" {
		context, err := sdkUtils.ReadAuthContext(configPath, cfContext)
		if err != nil {
			return err
		}
		cfAPIHost = context.URL
		cfAPIToken = context.Token
		logger.Debug("Using codefresh context", "Context-Name", context.Name, "Host", cfAPIHost)
	} else {
		logger.Debug("Reading credentials from environment variables")
		if cfAPIHost == "" {
			cfAPIHost = "https://g.codefresh.io"
		}
	}

	logger.Debug("Creating codefresh client", "host", cfAPIHost, "token", cfAPIToken)

	client := codefresh.New(&codefresh.ClientOptions{
		Auth: codefresh.AuthOptions{
			Token: cfAPIToken,
		},
		Host: cfAPIHost,
	})
	s.CodefreshAPI = &store.CodefreshAPI{
		Host:   cfAPIHost,
		Token:  cfAPIToken,
		Client: client,
	}

	return nil
}

func extendStoreWithKubeClient(logger logger.Logger) {
	s := store.GetStore()
	if kubeConfigPath == "" {
		currentUser, _ := user.Current()
		if currentUser != nil {
			kubeConfigPath = filepath.Join(currentUser.HomeDir, ".kube", "config")
			logger.Debug("Path to kubeconfig not set, using:", "kubeconfig", kubeConfigPath)
		}
	}

	s.KubernetesAPI = &store.KubernetesAPI{
		ConfigPath: kubeConfigPath,
	}
}

func setVerbosity(verbose bool) {
	s := store.GetStore()
	s.Verbose = verbose
}

func setInsecure(insecure bool) {
	s := store.GetStore()
	s.Insecure = insecure
}

func isUsingDefaultStorageClass(sc string) bool {
	if sc == "" {
		return true
	}
	return strings.HasPrefix(sc, plugins.DefaultStorageClassNamePrefix)
}

func dieOnError(err error) {
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func createTable() *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowLine(false)
	table.SetHeaderLine(false)
	table.SetColumnSeparator(" ")
	table.SetColWidth(100)
	return table
}

func getKubeClientBuilder(context string, namespace string, path string, inCluster bool) kube.Kube {
	return kube.New(&kube.Options{
		ContextName:      context,
		Namespace:        namespace,
		PathToKubeConfig: path,
		InCluster:        inCluster,
	})
}

func createLogger(command string, verbose bool, logFormatter string) logger.Logger {
	logFile := "venonalog.json"
	os.Remove(logFile)
	return logger.New(&logger.Options{
		Command:      command,
		Verbose:      verbose,
		LogToFile:    logFile,
		LogFormatter: logFormatter,
	})
}

func createSpinner(prefix, suffix string) *spinner.Spinner {
	s := spinner.New([]string{"   ", ".  ", ".. ", "..."}, 520*time.Millisecond)
	s.Suffix = suffix
	s.Prefix = prefix
	return s
}

func loadTolerationsFromFile(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		dieOnError(err)
	}

	return string(data)
}

func parseTolerations(s string) ([]k8sApi.Toleration, error) {
	if s == "" {
		return nil, nil
	}
	var data []k8sApi.Toleration
	err := json.Unmarshal([]byte(s), &data)
	if err != nil {
		return nil, fmt.Errorf("can not parse tolerations: %s", err)
	}
	return data, err
}

func fillKubernetesAPI(lgr logger.Logger, context string, namespace string, inCluster bool) {

	s := store.GetStore()
	if context == "" {
		config := clientcmd.GetConfigFromFileOrDie(s.KubernetesAPI.ConfigPath)
		context = config.CurrentContext
		lgr.Debug("Kube Context is not set, using current context", "Kube-Context-Name", context)
	}
	if namespace == "" {
		namespace = "default"
	}

	s.KubernetesAPI.InCluster = inCluster
	s.KubernetesAPI.ContextName = context
	s.KubernetesAPI.Namespace = namespace

}

func extendStoreWithAgentAPI(logger logger.Logger, token string, agentID string) {
	s := store.GetStore()
	logger.Debug("Using agent's token", "Token", token)
	s.AgentAPI = &store.AgentAPI{
		Token: token,
		Id:    agentID,
	}
}

// Parsing helpers --set-value , --set-file
// by https://github.com/helm/helm/blob/ec1d1a3d3eb672232f896f9d3b3d0797e4f519e3/pkg/cli/values/options.go#L41

// templateValuesToMap - processes cmd --values <values-file.yaml> --set-value k=v --set-file v=<context-of-file>
// using helm libraries
func templateValuesToMap(templateValueFiles, templateValues, templateFileValues []string) (map[string]interface{}, error) {
	valueOpts := &cliValues.Options{}
	if len(templateValueFiles) > 0 {
		for _, v := range templateValueFiles {
			valueOpts.ValueFiles = append(valueOpts.ValueFiles, v)
		}
	}

	if len(templateValues) > 0 {
		for _, v := range templateValues {
			valueOpts.Values = append(valueOpts.Values, v)
		}
	}

	if len(templateFileValues) > 0 {
		for _, v := range templateFileValues {
			valueOpts.FileValues = append(valueOpts.FileValues, v)
		}
	}
	valuesMap, err := valueOpts.MergeValues(getter.Providers{})
	return valuesMap, err
}

func mergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = mergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

// mergeValueStr - for merging cli parameters with mapped parameters
func mergeValueStr(valuesMap map[string]interface{}, key string, param *string, defaultValue ...string) {
	mapX := objx.New(valuesMap)
	if param != nil && *param != "" {
		mapX.Set(key, *param)
		return
	}
	val := mapX.Get(key).Str(defaultValue...)
	*param = val
}

// mergeValueBool - for merging cli parameters with mapped parameters
func mergeValueBool(valuesMap map[string]interface{}, key string, param *bool) {
	mapX := objx.New(valuesMap)
	if param != nil || *param == true {
		mapX.Set(key, *param)
		return
	}
	val := mapX.Get(key).Bool()
	*param = val
}

func mergeValueMSI(valuesMap map[string]interface{}, key string, param *map[string]interface{}, defaultValue ...map[string]interface{}) {
	mapX := objx.New(valuesMap)
	if param != nil && len(*param) > 0 {
		mapX.Set(key, *param)
		return
	}
	val := mapX.Get(key).MSI(defaultValue...)
	*param = val
}
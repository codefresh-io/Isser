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
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	// import all cloud providers auth clients
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/Masterminds/sprig"
	"github.com/codefresh-io/venona/venonactl/pkg/logger"
	templates "github.com/codefresh-io/venona/venonactl/pkg/templates/kubernetes"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

func unescape(s string) template.HTML {
	return template.HTML(s)
}

// template function to parse values for nodeSelector in form "key1=value1,key2=value2"
func nodeSelectorParamToYaml(ns string) string {
	nodeSelectorParts := strings.Split(ns, ",")
	var nodeSelectorYaml string
	for _, p := range(nodeSelectorParts){
		pSplit := strings.Split(p, "=")
		if len(pSplit) != 2 {
			continue
		}

		if len(nodeSelectorYaml) > 0 {
			nodeSelectorYaml += "\n"
		}
		nodeSelectorYaml += fmt.Sprintf("%s: %q", pSplit[0], pSplit[1])
	}
	return nodeSelectorYaml
}

// ExecuteTemplate - executes templates in tpl str with config as values
func ExecuteTemplate(tplStr string, data interface{}) (string, error) {
	funcMap := template.FuncMap{
		          "unescape": unescape,
							"nodeSelectorParamToYaml": nodeSelectorParamToYaml,
             }
	template, err := template.New("base").Funcs(sprig.FuncMap()).Funcs(funcMap).Parse(tplStr)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBufferString("")
	err = template.Execute(buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ParseTemplates - parses and exexute templates and return map of strings with obj data
func ParseTemplates(templatesMap map[string]string, data interface{}, pattern string, logger logger.Logger) (map[string]string, error) {
	parsedTemplates := make(map[string]string)
	nonEmptyParsedTemplateFunc := regexp.MustCompile(`[a-zA-Z0-9]`).MatchString
	for n, tpl := range templatesMap {
		match, _ := regexp.MatchString(pattern, n)
		if match != true {
			logger.Debug("Skipping parsing, pattern does not match", "Pattern", pattern, "Name", n)
			continue
		}
		logger.Debug("parsing template", "Name", n)
		tplEx, err := ExecuteTemplate(tpl, data)
		if err != nil {
			logger.Error("Failed to parse and execute template", "Name", n)
			return nil, err
		}

		// we add only non-empty parsedTemplates
		if nonEmptyParsedTemplateFunc(tplEx) {
			parsedTemplates[n] = tplEx
		}
	}
	return parsedTemplates, nil
}

// KubeObjectsFromTemplates return map of runtime.Objects from templateMap
// see https://github.com/kubernetes/client-go/issues/193 for examples
func KubeObjectsFromTemplates(templatesMap map[string]string, data interface{}, pattern string, logger logger.Logger) (map[string]runtime.Object, error) {
	parsedTemplates, err := ParseTemplates(templatesMap, data, pattern, logger)
	if err != nil {
		return nil, err
	}

	// Deserializing all kube objects from parsedTemplates
	// see https://github.com/kubernetes/client-go/issues/193 for examples
	kubeDecode := scheme.Codecs.UniversalDeserializer().Decode
	kubeObjects := make(map[string]runtime.Object)
	for n, objStr := range parsedTemplates {
		logger.Debug("Deserializing template", "Name", n)
		obj, groupVersionKind, err := kubeDecode([]byte(objStr), nil, nil)
		if err != nil {
			logger.Error(fmt.Sprintf("Cannot deserialize kuberentes object %s: %v", n, err))
			return nil, err
		}
		logger.Debug("deserializing template success", "Name", n, "Group", groupVersionKind.Group)
		kubeObjects[n] = obj
	}
	return kubeObjects, nil
}

func getKubeObjectsFromTempalte(values map[string]interface{}, pattern string, logger logger.Logger) (map[string]runtime.Object, error) {
	templatesMap := templates.TemplatesMap()
	return KubeObjectsFromTemplates(templatesMap, values, pattern, logger)
}

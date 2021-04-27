module github.com/codefresh-io/go/venona

go 1.15

require (
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/go-retryablehttp v0.6.7
	github.com/inconshreveable/log15 v0.0.0-20200109203555-b30bc20e4fd1
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/newrelic/go-agent/v3 v3.10.0
	github.com/newrelic/go-agent/v3/integrations/nrgorilla v1.1.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.1
	github.com/stretchr/objx v0.2.0
	github.com/stretchr/testify v1.6.1
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4

)

replace github.com/bketelsen/crypt => github.com/bketelsen/crypt v0.0.3

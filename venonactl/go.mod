module github.com/codefresh-io/venona/venonactl

go 1.15

require (
	github.com/Masterminds/semver v1.5.0
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/briandowns/spinner v1.12.0
	github.com/codefresh-io/go-sdk v0.24.0
	github.com/dustin/go-humanize v1.0.0
	github.com/inconshreveable/log15 v0.0.0-20201112154412-8562bdadbbac
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/olekukonko/tablewriter v0.0.5
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	github.com/stretchr/objx v0.3.0
	golang.org/x/sys v0.0.0-20210320140829-1e4c9ba3b0c4 // indirect
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.5.3
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/client-go v0.20.4
)

replace (
	github.com/bketelsen/crypt => github.com/bketelsen/crypt v0.0.3
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)

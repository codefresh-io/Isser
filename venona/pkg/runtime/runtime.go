package runtime

import (
	"github.com/codefresh-io/go/venona/pkg/codefresh"
	"github.com/codefresh-io/go/venona/pkg/kubernetes"
)
 

type (
	// Runtime API client
	Runtime interface {
		StartWorkflow([]codefresh.Task) error
		TerminateWorkflow([]codefresh.Task) error
	}

	Options struct{
		Kubernetes kubernetes.Kubernetes
	}

	runtime struct{}
)

// New creates new Runtime client
func New(opt Options) (Runtime) {
	return &runtime{}
}

func (r runtime) StartWorkflow(tasks []codefresh.Task) error {
	return nil
}
func (r runtime) TerminateWorkflow(tasks []codefresh.Task) error {
	return nil
}

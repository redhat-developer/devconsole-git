package log

import (
	"github.com/go-logr/logr"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
)

type GitSourceLogger struct {
	logr.Logger
}

func LogWithGSValues(log logr.Logger, gitSource *v1alpha1.GitSource, additional ...interface{}) *GitSourceLogger {
	var values []interface{}
	if gitSource != nil {
		values = []interface{}{
			"name", gitSource.ObjectMeta.Name,
			"url", gitSource.Spec.URL,
			"ref", gitSource.Spec.Ref,
			"flavor", gitSource.Spec.Flavor,
		}
	}
	return &GitSourceLogger{Logger: log.WithValues(append(values, additional...)...)}
}

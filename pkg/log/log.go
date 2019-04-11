package log

import (
	"github.com/go-logr/logr"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
)

func LogWithGSValues(log logr.Logger, gitSource *v1alpha1.GitSource, additional ...interface{}) logr.Logger {
	values := []interface{}{
		"name", gitSource.ObjectMeta.Name,
		"url", gitSource.Spec.URL,
		"ref", gitSource.Spec.Ref,
		"flavor", gitSource.Spec.Flavor,
	}
	values = append(values, additional...)

	return log.WithValues(values...)
}

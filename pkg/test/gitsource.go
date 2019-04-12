package test

import (
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	GitSourceName         = "some-repo-gs"
	GitSourceAnalysisName = "some-repo-gsa"
	Namespace             = "my-namespace"
)

type GitSourceModifier func(*v1alpha1.GitSource)

func WithFlavor(flavor string) GitSourceModifier {
	return func(gitSource *v1alpha1.GitSource) {
		gitSource.Spec.Flavor = flavor
	}
}

func WithURL(url string) GitSourceModifier {
	return func(gitSource *v1alpha1.GitSource) {
		gitSource.Spec.URL = url
	}
}

func WithRef(ref string) GitSourceModifier {
	return func(gitSource *v1alpha1.GitSource) {
		gitSource.Spec.Ref = ref
	}
}

func NewGitSource(modifiers ...GitSourceModifier) *v1alpha1.GitSource {
	gitSource := &v1alpha1.GitSource{
		ObjectMeta: v1.ObjectMeta{
			Name:      GitSourceName,
			Namespace: Namespace,
		},
		Spec: v1alpha1.GitSourceSpec{},
	}
	for _, modify := range modifiers {
		modify(gitSource)
	}
	return gitSource
}

func NewGitSourceAnalysis(gitSourceName string) *v1alpha1.GitSourceAnalysis {
	gsa := &v1alpha1.GitSourceAnalysis{
		ObjectMeta: v1.ObjectMeta{
			Name:      GitSourceAnalysisName,
			Namespace: Namespace,
		},
		Spec: v1alpha1.GitSourceAnalysisSpec{
			GitSourceRef: v1alpha1.GitSourceRef{
				Name: gitSourceName,
			},
		},
	}
	return gsa
}

package test

import "github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"

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
		Spec: v1alpha1.GitSourceSpec{},
	}
	for _, modify := range modifiers {
		modify(gitSource)
	}
	return gitSource
}

package repository

import (
	"github.com/redhat-developer/git-service/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/git-service/pkg/git"
)

type GitService interface {
	GetListOfFilesInRootDir() ([]string, error)
	GetLanguageList() ([]string, error)
}

type ServiceCreator func(gitSource *v1alpha1.GitSource, secret git.Secret) (GitService, error)

func NewGitService(gitSource *v1alpha1.GitSource, secret git.Secret, serviceCreators []ServiceCreator) (GitService, error) {

	for _, creator := range serviceCreators {
		detector, err := creator(gitSource, secret)
		if err != nil {
			return nil, err
		}
		if detector != nil {
			return detector, nil
		}
	}

	return nil, nil
}

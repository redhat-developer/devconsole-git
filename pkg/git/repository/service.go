package repository

import (
	"github.com/redhat-developer/git-service/pkg/git"
)

type GitService interface {
	GetListOfFilesInRootDir() ([]string, error)
	GetLanguageList() ([]string, error)
}

type ServiceCreator func(gitSource *git.Source) (GitService, error)

func NewGitService(gitSource *git.Source, serviceCreators []ServiceCreator) (GitService, error) {

	for _, creator := range serviceCreators {
		detector, err := creator(gitSource)
		if err != nil {
			return nil, err
		}
		if detector != nil {
			return detector, nil
		}
	}

	return nil, nil
}

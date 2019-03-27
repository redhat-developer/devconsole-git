package repository

import (
	"github.com/redhat-developer/git-service/pkg/git"
)

type Service interface {
	GetListOfFilesInRootDir() ([]string, error)
	GetLanguageList() ([]string, error)
}

type ServiceCreator func(gitSource *git.Source) (Service, error)

func NewService(gitSource *git.Source, serviceCreators []ServiceCreator) (Service, error) {

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

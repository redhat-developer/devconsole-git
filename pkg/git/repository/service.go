package repository

import (
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/git-service/pkg/git"
)

type GitService interface {
	// GetListOfFilesInRootDir returns list of filenames present in the root directory
	GetListOfFilesInRootDir() ([]string, error)
	// GetLanguageList returns list of detected languages in the sorted order where the first one is the most used
	GetLanguageList() ([]string, error)
}

// ServiceCreator creates an instance of GitService for the given v1alpha1.GitSource
type ServiceCreator func(gitSource *v1alpha1.GitSource, secret *git.SecretProvider) (GitService, error)

// NewGitService returns an instance of GitService for the given v1alpha1.GitSource. If no service is matched then returns nil
func NewGitService(gitSource *v1alpha1.GitSource, secret *git.SecretProvider, serviceCreators []ServiceCreator) (GitService, error) {

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

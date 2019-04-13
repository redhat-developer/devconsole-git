package repository

import (
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/redhat-developer/git-service/pkg/git/detector/build"
	"github.com/redhat-developer/git-service/pkg/log"
)

type GitService interface {
	// FileExistenceChecker returns an instance of checker for existence of files in the root directory
	FileExistenceChecker() (FileExistenceChecker, error)
	// GetLanguageList returns list of detected languages in the sorted order where the first one is the most used
	GetLanguageList() ([]string, error)
}

type FileExistenceChecker interface {
	// GetListOfFoundFiles returns list of filenames present in the root directory
	GetListOfFoundFiles() []string
	// DetectFiles detects if any of the build tool files are present in the root directory
	DetectFiles(buildTool build.Tool) []string
}

// ServiceCreator creates an instance of GitService for the given v1alpha1.GitSource
type ServiceCreator func(log *log.GitSourceLogger, gitSource *v1alpha1.GitSource, secret *git.SecretProvider) (GitService, error)

// NewGitService returns an instance of GitService for the given v1alpha1.GitSource. If no service is matched then returns nil
func NewGitService(log *log.GitSourceLogger, gitSource *v1alpha1.GitSource, secret *git.SecretProvider, serviceCreators []ServiceCreator) (GitService, error) {

	for _, creator := range serviceCreators {
		detector, err := creator(log, gitSource, secret)
		if err != nil {
			return nil, err
		}
		if detector != nil {
			return detector, nil
		}
	}

	return nil, nil
}

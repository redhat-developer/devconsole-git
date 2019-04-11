package detector

import (
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/redhat-developer/git-service/pkg/git/detector/build"
	"github.com/redhat-developer/git-service/pkg/git/repository"
	"github.com/redhat-developer/git-service/pkg/git/repository/bitbucket"
	"github.com/redhat-developer/git-service/pkg/git/repository/gitlab"
	"sync"

	"github.com/redhat-developer/git-service/pkg/git/repository/generic"
	"github.com/redhat-developer/git-service/pkg/git/repository/github"
)

var gitServiceCreators = []repository.ServiceCreator{
	github.NewRepoServiceIfMatches(),
	bitbucket.NewRepoServiceIfMatches(),
	gitlab.NewRepoServiceIfMatches(),
}

// DetectBuildEnvironmentsWithSecret detects build tools and languages using the given secret in the git repository
// defined by the given v1alpha1.GitSource
func DetectBuildEnvironmentsWithSecret(gitSource *v1alpha1.GitSource, secret git.Secret) (*v1alpha1.BuildEnvStats, error) {
	return detectBuildEnvs(gitSource, git.NewSecretProvider(secret), gitServiceCreators)
}

// DetectBuildEnvironments detects build tools and languages using the default secret in the git repository
// defined by the given v1alpha1.GitSource
func DetectBuildEnvironments(gitSource *v1alpha1.GitSource) (*v1alpha1.BuildEnvStats, error) {
	return detectBuildEnvs(gitSource, git.NewSecretProvider(nil), gitServiceCreators)
}

func detectBuildEnvs(gitSource *v1alpha1.GitSource, secretProvider *git.SecretProvider, serviceCreators []repository.ServiceCreator) (*v1alpha1.BuildEnvStats, error) {
	service, err := repository.NewGitService(gitSource, secretProvider, serviceCreators)
	if err != nil {
		return nil, err
	}
	if service == nil {
		service, err = generic.NewRepositoryService(gitSource, secretProvider)
		if err != nil {
			return nil, err
		}
	}
	return detectBuildEnvsUsingService(service)
}

func detectBuildEnvsUsingService(service repository.GitService) (*v1alpha1.BuildEnvStats, error) {
	var wg sync.WaitGroup
	wg.Add(1)
	detectedBuildTools := make(chan *v1alpha1.DetectedBuildType, len(build.Tools))
	var detectionErr error
	go func() {
		defer wg.Done()
		detectionErr = detectBuildTools(service, detectedBuildTools)
	}()

	languageList, err := service.GetLanguageList()
	if err != nil {
		return nil, err
	}
	wg.Wait()
	if detectionErr != nil {
		return nil, detectionErr
	}

	var environments []v1alpha1.DetectedBuildType
	for detectedBuildTool := range detectedBuildTools {
		if detectedBuildTool != nil {
			environments = append(environments, *detectedBuildTool)
		}
	}

	return &v1alpha1.BuildEnvStats{
		SortedLanguages:    languageList,
		DetectedBuildTypes: environments,
	}, nil
}

func detectBuildTools(service repository.GitService, detectedBuildTools chan *v1alpha1.DetectedBuildType) error {
	var wg sync.WaitGroup
	wg.Add(len(build.Tools))

	fileExistenceChecker, err := service.FileExistenceChecker()
	if err != nil {
		return err
	}

	for _, tool := range build.Tools {
		go func(buildTool build.Tool) {
			defer wg.Done()
			detectedFiles := fileExistenceChecker.DetectFiles(buildTool)
			if len(detectedFiles) > 0 {
				detectedBuildTools <- build.NewDetectedBuildTool(buildTool.Language, buildTool.Name, detectedFiles)
			}
		}(tool)
	}

	wg.Wait()
	close(detectedBuildTools)
	return nil
}

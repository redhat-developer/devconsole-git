package detector

import (
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/redhat-developer/git-service/pkg/git/repository"
	"github.com/redhat-developer/git-service/pkg/git/repository/bitbucket"
	"github.com/redhat-developer/git-service/pkg/git/repository/gitlab"
	"regexp"
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
	return DetectBuildEnvironments(gitSource, git.NewSecretProvider(secret))
}

// DetectBuildEnvironments detects build tools and languages using the secret provided by the SecretProvider
// in the git repository defined by the given v1alpha1.GitSource
func DetectBuildEnvironments(gitSource *v1alpha1.GitSource, secretProvider *git.SecretProvider) (*v1alpha1.BuildEnvStats, error) {
	return detectBuildEnvs(gitSource, secretProvider, gitServiceCreators)
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
	detectedBuildTools := make(chan *v1alpha1.DetectedBuildType, len(BuildTools))
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
	wg.Add(len(BuildTools))

	files, err := service.GetListOfFilesInRootDir()
	if err != nil {
		return err
	}

	for _, tool := range BuildTools {
		go func(buildTool BuildTool) {
			defer wg.Done()
			detectedFiles := detectBuildToolFiles(buildTool, files)
			if len(detectedFiles) > 0 {
				detectedBuildTools <- NewDetectedBuildTool(buildTool.Language, buildTool.Name, detectedFiles)
			}
		}(tool)
	}

	wg.Wait()
	close(detectedBuildTools)
	return nil
}

func detectBuildToolFiles(buildTool BuildTool, filenames []string) []string {
	detectedFiles := make(chan string, len(buildTool.ExpectedFiles))
	var wg sync.WaitGroup
	wg.Add(len(buildTool.ExpectedFiles))

	for _, file := range buildTool.ExpectedFiles {
		go func(buildToolFile *regexp.Regexp) {
			defer wg.Done()
			if detectedFile := getFileIfExists(buildToolFile, filenames); detectedFile != "" {
				detectedFiles <- detectedFile
			}
		}(file)
	}

	wg.Wait()
	close(detectedFiles)
	var result []string
	for detectedFile := range detectedFiles {
		if detectedFile != "" {
			result = append(result, detectedFile)
		}
	}

	return result
}

func getFileIfExists(buildToolFile *regexp.Regexp, actualFiles []string) string {
	for _, file := range actualFiles {
		if buildToolFile.MatchString(file) {
			return file
		}
	}
	return ""
}

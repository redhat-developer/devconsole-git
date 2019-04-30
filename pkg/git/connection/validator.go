package connection

import (
	"fmt"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/devconsole-git/pkg/git"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository/bitbucket"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository/generic"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository/github"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository/gitlab"
	"github.com/redhat-developer/devconsole-git/pkg/log"
	gittransport "gopkg.in/src-d/go-git.v4/plumbing/transport"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

var gitServiceCreators = []repository.ServiceCreator{
	github.NewRepoServiceIfMatches(),
	bitbucket.NewRepoServiceIfMatches(),
	gitlab.NewRepoServiceIfMatches(),
}

// ValidateGitSource validates if a git repository defined by the given GitSource is reachable
// and if it contains the defined branch (master if empty)
func ValidateGitSource(log *log.GitSourceLogger, gitSource *v1alpha1.GitSource) *ValidationError {
	endpoint, err := gittransport.NewEndpoint(gitSource.Spec.URL)
	if err != nil {
		return newValidationErrorf(v1alpha1.RepoNotReachable, "unable to parse the URL: %s", err.Error())
	}
	if endpoint.Host == "" {
		return newValidationErrorf(v1alpha1.RepoNotReachable, "the URL doesn't contain host")
	}
	client := &http.Client{}
	path := endpoint.Path
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	if !strings.HasSuffix(path, ".git") {
		path += ".git"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	url := fmt.Sprintf("https://%s%s/info/refs?service=git-upload-pack", endpoint.Host, path)
	resp, err := client.Get(url)
	if err != nil {
		return newValidationErrorf(v1alpha1.RepoNotReachable, err.Error())
	}
	return validateBranch(log, gitSource.Spec.Ref, resp)
}

func validateBranch(log *log.GitSourceLogger, branch string, resp *http.Response) *ValidationError {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err, "error while reading body")
		return newValidationErrorf(v1alpha1.RepoNotReachable, err.Error())
	}
	err = resp.Body.Close()
	if err != nil {
		log.Error(err, "error while closing body")
		return newValidationErrorf(v1alpha1.RepoNotReachable, err.Error())
	}

	if resp.StatusCode != http.StatusOK {
		return newValidationErrorf(v1alpha1.RepoNotReachable, "server responded with %s", resp.Status)
	}
	if branch == "" {
		branch = repository.Master
	}
	branchRefRegexp := fmt.Sprintf(`\ refs\/heads\/%s\n`, branch)
	compile, err := regexp.Compile(branchRefRegexp)
	if err != nil {
		log.Error(err, "error while parsing regexp of the given branch")
		return newValidationErrorf(v1alpha1.BranchNotFound, "cannot parse the branch: %s", err.Error())
	}
	if compile.Find(body) != nil {
		return nil
	}
	return newValidationErrorf(v1alpha1.BranchNotFound, "cannot find the branch")
}

// ValidateGitSourceWithSecret detects build tools and languages using the given secret in the git repository
// defined by the given v1alpha1.GitSource
func ValidateGitSourceWithSecret(log *log.GitSourceLogger, gitSource *v1alpha1.GitSource, secret git.Secret) *ValidationError {
	return validateGitSourceWithSecret(log, gitSource, git.NewSecretProvider(secret), gitServiceCreators)
}

func validateGitSourceWithSecret(log *log.GitSourceLogger,
	gitSource *v1alpha1.GitSource,
	secretProvider *git.SecretProvider,
	serviceCreators []repository.ServiceCreator) *ValidationError {

	service, err := repository.NewGitService(log, gitSource, secretProvider, serviceCreators)
	if err != nil {
		return newValidationErrorf(v1alpha1.RepoNotReachable, err.Error())
	}
	if service == nil {
		service, err = generic.NewRepositoryService(gitSource, secretProvider)
		if err != nil {
			return newValidationErrorf(v1alpha1.RepoNotReachable, err.Error())
		}
	}
	if err := service.CheckCredentials(); err != nil {
		return newValidationErrorf(v1alpha1.BadCredentials, "cannot get user information: %s", err.Error())
	}
	if err := service.CheckRepoAccessibility(); err != nil {
		return newValidationErrorf(v1alpha1.RepoNotReachable, "unable to reach the URL: %s", err.Error())
	}
	if err := service.CheckBranch(); err != nil {
		return newValidationErrorf(v1alpha1.BranchNotFound, "unable to find the branch: %s", err.Error())
	}
	return nil
}

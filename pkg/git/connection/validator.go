package connection

import (
	"fmt"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository"
	"github.com/redhat-developer/devconsole-git/pkg/log"
	gittransport "gopkg.in/src-d/go-git.v4/plumbing/transport"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

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

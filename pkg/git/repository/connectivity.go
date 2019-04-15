package repository

import (
	"fmt"
	"github.com/redhat-developer/devconsole-git/pkg/log"
	gittransport "gopkg.in/src-d/go-git.v4/plumbing/transport"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const (
	unableReachRepoError   = "unable to reach the repo"
	unableFindBranchError  = "unable to find the branch"
	unableParseBranchError = "unable to parse the branch"
)

func IsReachableWithBranch(log *log.GitSourceLogger, branch string, endpoint *gittransport.Endpoint) (bool, error) {
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
	if err != nil || endpoint.Host == "" || endpoint.Path == "" {
		return false, fmt.Errorf(unableReachRepoError)
	} else {
		return validateBranch(log, branch, resp)
	}
}

func validateBranch(log *log.GitSourceLogger, branch string, resp *http.Response) (bool, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err, "error while reading body")
		return false, fmt.Errorf(unableReachRepoError)
	}
	err = resp.Body.Close()
	if err != nil {
		log.Error(err, "error while closing body")
		return false, fmt.Errorf(unableReachRepoError)
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf(unableReachRepoError)
	} else {
		if branch == "" {
			branch = Master
		}
		branchRefRegexp := fmt.Sprintf(`\ refs\/heads\/%s\n`, branch)
		compile, err := regexp.Compile(branchRefRegexp)
		if err != nil {
			log.Error(err, "error while parsing regexp of the given branch")
			return false, fmt.Errorf(unableParseBranchError)
		} else {
			if compile.Find(body) != nil {
				return true, nil
			} else {
				return false, fmt.Errorf(unableFindBranchError)
			}
		}
	}
}

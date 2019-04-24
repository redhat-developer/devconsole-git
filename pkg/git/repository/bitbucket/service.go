package bitbucket

import (
	"encoding/json"
	"fmt"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/devconsole-git/pkg/git"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository"
	"github.com/redhat-developer/devconsole-git/pkg/log"
	gittransport "gopkg.in/src-d/go-git.v4/plumbing/transport"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	bitbucketHost   = "bitbucket.org"
	bitbucketFlavor = "bitbucket"
)

type RepositoryService struct {
	secret  git.Secret
	client  *http.Client
	baseURL string
	repo    repository.StructuredIdentifier
	log     *log.GitSourceLogger
}

func NewRepoServiceIfMatches() repository.ServiceCreator {
	return func(log *log.GitSourceLogger, gitSource *v1alpha1.GitSource, secretProvider *git.SecretProvider) (repository.GitService, error) {
		if secretProvider.SecretType() == git.SshKeyType {
			return nil, nil
		}
		endpoint, err := gittransport.NewEndpoint(gitSource.Spec.URL)
		if err != nil {
			return nil, err
		}
		if endpoint.Host == bitbucketHost || gitSource.Spec.Flavor == bitbucketFlavor {
			secret := secretProvider.GetSecret(git.NewOauthToken([]byte("")))
			return newBbService(log, gitSource, endpoint, secret)
		}
		return nil, nil
	}
}

func newBbService(log *log.GitSourceLogger, gitSource *v1alpha1.GitSource, endpoint *gittransport.Endpoint, secret git.Secret) (*RepositoryService, error) {
	repo, err := repository.NewStructuredIdentifier(gitSource, endpoint)
	if err != nil {
		return nil, err
	}
	client := secret.Client()

	return &RepositoryService{
		secret:  secret,
		client:  client,
		repo:    repo,
		baseURL: getBaseURL(endpoint),
		log:     log,
	}, nil
}

func getBaseURL(endpoint *gittransport.Endpoint) string {
	baseURL := "https://api.bitbucket.org/"
	if endpoint.Host != bitbucketHost {
		host := endpoint.Host
		if !strings.HasPrefix(host, "api.") {
			host = "api." + host
		}
		protocol := endpoint.Protocol
		if endpoint.Protocol == "ssh" || endpoint.Protocol == "git" {
			protocol = "https"
		}
		baseURL = fmt.Sprintf("%s://%s/", protocol, host)
	}
	return baseURL
}

func (s *RepositoryService) FileExistenceChecker() (repository.FileExistenceChecker, error) {
	apiURL := fmt.Sprintf(`%s2.0/repositories/%s/%s/src/%s/?q=type="commit_file"`,
		s.baseURL, s.repo.Owner, s.repo.Name, s.repo.Branch)
	files, err := s.doPaginatedCalls(apiURL)
	if err != nil {
		return nil, err
	}
	return repository.NewCheckerWithFetchedFiles(files), nil
}

func (s *RepositoryService) doPaginatedCalls(apiURL string) ([]string, error) {
	respBody, err := s.do(apiURL)
	if err != nil {
		return nil, err
	}
	var files Src
	err = json.Unmarshal(respBody, &files)
	if err != nil {
		return nil, err
	}
	var filenames []string
	for _, entry := range files.Values {
		filenames = append(filenames, entry.Path)
	}
	if files.Next != "" {
		anotherFiles, err := s.doPaginatedCalls(files.Next)
		if err != nil {
			return nil, err
		}
		filenames = append(filenames, anotherFiles...)
	}
	return filenames, nil
}

func (s *RepositoryService) GetLanguageList() ([]string, error) {
	apiURL := fmt.Sprintf(`%s2.0/repositories/%s/%s/`, s.baseURL, s.repo.Owner, s.repo.Name)
	respBody, err := s.do(apiURL)
	if err != nil {
		return nil, err
	}
	var repoLanguage RepositoryLanguage
	err = json.Unmarshal(respBody, &repoLanguage)
	if err != nil {
		return nil, err
	}

	return []string{repoLanguage.Language}, nil
}

func (s *RepositoryService) CheckCredentials() error {
	apiURL := fmt.Sprintf(`%s2.0/user`, s.baseURL)
	_, err := s.do(apiURL)
	return err
}

func (s *RepositoryService) CheckRepoAccessibility() error {
	apiURL := fmt.Sprintf(`%s2.0/repositories/%s/%s/`, s.baseURL, s.repo.Owner, s.repo.Name)
	_, err := s.do(apiURL)
	return err
}

func (s *RepositoryService) CheckBranch() error {
	apiURL := fmt.Sprintf(`%s2.0/repositories/%s/%s/refs/branches/%s`, s.baseURL, s.repo.Owner, s.repo.Name, s.repo.Branch)
	_, err := s.do(apiURL)
	return err
}

func (s *RepositoryService) do(apiURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	if s.secret.SecretType() == git.UsernamePasswordType {
		req.SetBasicAuth(git.ParseUsernameAndPassword(s.secret.SecretContent()))
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			s.log.Error(err, "closing body failed")
		}
	}()

	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		errMsg := "call to the API endpoint %s failed with [%s] and message [%s]"
		var respErr ResponseError
		err = json.Unmarshal(respBody, &respErr)
		if err != nil || respErr.Error.Message == "" {
			return nil, fmt.Errorf(errMsg, apiURL, resp.Status, string(respBody))
		}
		return nil, fmt.Errorf(errMsg, apiURL, resp.Status, respErr.Error.Message)
	}
	return respBody, nil
}

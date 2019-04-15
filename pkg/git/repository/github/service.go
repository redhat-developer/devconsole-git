package github

import (
	"context"
	"fmt"
	gogh "github.com/google/go-github/github"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/devconsole-git/pkg/git"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository"
	"github.com/redhat-developer/devconsole-git/pkg/log"
	gittransport "gopkg.in/src-d/go-git.v4/plumbing/transport"
)

const (
	githubHost   = "github.com"
	githubFlavor = "github"
)

var anonymousSecret = git.NewUsernamePassword("anonymous", "")

type RepositoryService struct {
	gitSource *v1alpha1.GitSource
	client    *gogh.Client
	repo      repository.StructuredIdentifier
	filenames []string
	secret    git.Secret
	log       *log.GitSourceLogger
}

// NewRepoServiceIfMatches returns function creating Github repository service if either host of the git repo URL is github.com
// or flavor of the given git source is github then, nil otherwise
func NewRepoServiceIfMatches() repository.ServiceCreator {
	return func(log *log.GitSourceLogger, gitSource *v1alpha1.GitSource, secretProvider *git.SecretProvider) (repository.GitService, error) {
		if secretProvider.SecretType() == git.SshKeyType {
			return nil, nil
		}
		endpoint, err := gittransport.NewEndpoint(gitSource.Spec.URL)
		if err != nil {
			return nil, err
		}
		if endpoint.Host == githubHost || gitSource.Spec.Flavor == githubFlavor {
			secret := secretProvider.GetSecret(anonymousSecret)
			return newGhService(log, gitSource, endpoint, secret)
		}
		return nil, nil
	}
}

func newGhService(log *log.GitSourceLogger, gitSource *v1alpha1.GitSource, endpoint *gittransport.Endpoint, secret git.Secret) (*RepositoryService, error) {
	repo, err := repository.NewStructuredIdentifier(gitSource, endpoint)
	if err != nil {
		return nil, err
	}
	baseClient := secret.Client()
	if secret.SecretType() == git.UsernamePasswordType {
		username, password := git.ParseUsernameAndPassword(secret.SecretContent())
		baseClient.Transport = &gogh.BasicAuthTransport{Username: username, Password: password}
	}
	client := gogh.NewClient(baseClient)

	return &RepositoryService{
		gitSource: gitSource,
		client:    client,
		repo:      repo,
		secret:    secret,
		log:       log,
	}, nil
}

func (s *RepositoryService) FileExistenceChecker() (repository.FileExistenceChecker, error) {
	if isAnonymousSecret(s.secret) {
		baseURL := fmt.Sprintf("https://github.com/%s/%s/blob/%s/", s.repo.Owner, s.repo.Name, s.repo.Branch)
		return repository.NewCheckerUsingHeaderRequests(s.log, baseURL, s.secret), nil
	}

	tree, _, err := s.client.Git.GetTree(
		context.Background(),
		s.repo.Owner,
		s.repo.Name,
		s.repo.Branch,
		false)
	if err != nil {
		return nil, err
	}
	var filenames []string
	for _, entry := range tree.Entries {
		filenames = append(filenames, *entry.Path)
	}
	return repository.NewCheckerWithFetchedFiles(filenames), nil
}

func (s *RepositoryService) GetLanguageList() ([]string, error) {
	if isAnonymousSecret(s.secret) {
		return []string{}, nil
	}

	languages, _, err := s.client.Repositories.ListLanguages(
		context.Background(),
		s.repo.Owner,
		s.repo.Name)

	if err != nil {
		return nil, err
	}

	return git.SortLanguagesWithInts(languages), nil
}

func isAnonymousSecret(secret git.Secret) bool {
	return secret.SecretType() == git.UsernamePasswordType &&
		secret.SecretContent() == anonymousSecret.SecretContent()
}

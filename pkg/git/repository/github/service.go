package github

import (
	"context"
	gogh "github.com/google/go-github/github"
	"github.com/redhat-developer/git-service/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/redhat-developer/git-service/pkg/git/repository"
	gittransport "gopkg.in/src-d/go-git.v4/plumbing/transport"
)

const (
	githubHost   = "github.com"
	githubFlavor = "github"
)

type RepositoryService struct {
	gitSource *v1alpha1.GitSource
	client    *gogh.Client
	repo      repository.StructuredIdentifier
	filenames []string
}

func NewRepoServiceIfMatches() repository.ServiceCreator {
	return func(gitSource *v1alpha1.GitSource, secret git.Secret) (repository.GitService, error) {
		endpoint, err := gittransport.NewEndpoint(gitSource.Spec.URL)
		if err != nil {
			return nil, err
		}
		if endpoint.Host == githubHost || gitSource.Spec.Flavor == githubFlavor {
			return newGhService(gitSource, endpoint, secret)
		}
		return nil, nil
	}
}

func newGhService(gitSource *v1alpha1.GitSource, endpoint *gittransport.Endpoint, secret git.Secret) (*RepositoryService, error) {
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
	}, nil
}

func (s *RepositoryService) GetListOfFilesInRootDir() ([]string, error) {
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
	return filenames, nil
}

func (s *RepositoryService) GetLanguageList() ([]string, error) {
	languages, _, err := s.client.Repositories.ListLanguages(
		context.Background(),
		s.repo.Owner,
		s.repo.Name)

	if err != nil {
		return nil, err
	}

	return git.SortLanguagesWithInts(languages), nil
}

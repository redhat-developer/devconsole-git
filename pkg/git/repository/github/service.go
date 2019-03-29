package github

import (
	"context"
	gogh "github.com/google/go-github/github"
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/redhat-developer/git-service/pkg/git/repository"
	gittransport "gopkg.in/src-d/go-git.v4/plumbing/transport"
)

const (
	githubHost   = "github.com"
	githubFlavor = "github"
)

type RepositoryService struct {
	gitSource *git.Source
	client    *gogh.Client
	repo      repository.StructuredIdentifier
	filenames []string
}

func NewRepoServiceIfMatches() repository.ServiceCreator {
	return func(gitSource *git.Source) (repository.GitService, error) {
		endpoint, err := gittransport.NewEndpoint(gitSource.URL)
		if err != nil {
			return nil, err
		}
		if endpoint.Host == githubHost || gitSource.Flavor == githubFlavor {
			return newGhService(gitSource, endpoint)
		}
		return nil, nil
	}
}

func newGhService(gitSource *git.Source, endpoint *gittransport.Endpoint) (*RepositoryService, error) {
	repo, err := repository.NewStructuredIdentifier(gitSource, endpoint)
	if err != nil {
		return nil, err
	}
	baseClient := gitSource.Secret.Client()
	if gitSource.Secret.SecretType() == git.UsernamePasswordType {
		username, password := git.ParseUsernameAndPassword(gitSource.Secret.SecretContent())
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

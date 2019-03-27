package gitlab

import (
	"github.com/redhat-developer/git-service/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/redhat-developer/git-service/pkg/git/repository"
	gogl "github.com/xanzy/go-gitlab"
	gittransport "gopkg.in/src-d/go-git.v4/plumbing/transport"
)

const (
	gitlabHost   = "gitlab.com"
	gitlabFlavor = "gitlab"
)

type RepositoryService struct {
	client *gogl.Client
	repo   repository.StructuredIdentifier
}

func NewRepoServiceIfMatches() repository.ServiceCreator {
	return func(gitSource *v1alpha1.GitSource, secret git.Secret) (repository.GitService, error) {
		endpoint, err := gittransport.NewEndpoint(gitSource.Spec.URL)
		if err != nil {
			return nil, err
		}

		if endpoint.Host == gitlabHost || gitSource.Spec.Flavor == gitlabFlavor {
			return newGhClient(gitSource, secret, endpoint)
		}
		return nil, nil
	}
}

func newGhClient(gitSource *v1alpha1.GitSource, secret git.Secret, endpoint *gittransport.Endpoint) (*RepositoryService, error) {
	repo, err := repository.NewStructuredIdentifier(gitSource, endpoint)
	if err != nil {
		return nil, err
	}

	client := gogl.NewOAuthClient(secret.Client(), secret.SecretContent())
	if secret.SecretType() == git.UsernamePasswordType {
		username, password := git.ParseUsernameAndPassword(secret.SecretContent())
		shortUrl := *endpoint
		shortUrl.Path = ""
		client, err = gogl.NewBasicAuthClient(secret.Client(), shortUrl.String(), username, password)
		if err != nil {
			return nil, err
		}
	}

	return &RepositoryService{
		client: client,
		repo:   repo,
	}, nil
}

func (s *RepositoryService) GetListOfFilesInRootDir() ([]string, error) {
	tree, _, err := s.client.Repositories.ListTree(
		s.repo.OwnerWithName(),
		&gogl.ListTreeOptions{
			Ref: &s.repo.Branch,
		})
	if err != nil {
		return nil, err
	}
	var filenames []string
	for _, entry := range tree {
		filenames = append(filenames, entry.Path)
	}
	return filenames, nil
}

func (s *RepositoryService) GetLanguageList() ([]string, error) {
	languages, _, err := s.client.Projects.GetProjectLanguages(s.repo.OwnerWithName())
	if err != nil {
		return nil, err
	}

	return git.SortLanguagesWithFloats32(*languages), nil
}

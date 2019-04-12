package gitlab

import (
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/redhat-developer/git-service/pkg/git/repository"
	"github.com/redhat-developer/git-service/pkg/log"
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

// NewRepoServiceIfMatches returns function creating Github repository service if either host of the git repo URL is gitlab.com
// or flavor of the given git source is gitlab then, nil otherwise
func NewRepoServiceIfMatches() repository.ServiceCreator {
	return func(log *log.GitSourceLogger, gitSource *v1alpha1.GitSource, secretProvider *git.SecretProvider) (repository.GitService, error) {
		if secretProvider.SecretType() == git.SshKeyType {
			return nil, nil
		}
		endpoint, err := gittransport.NewEndpoint(gitSource.Spec.URL)
		if err != nil {
			return nil, err
		}

		if endpoint.Host == gitlabHost || gitSource.Spec.Flavor == gitlabFlavor {
			secret := secretProvider.GetSecret(git.NewOauthToken([]byte("")))
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

	client := gogl.NewClient(nil, secret.SecretContent())
	client.SetBaseURL(getBaseUrl(endpoint))

	if secret.SecretType() == git.UsernamePasswordType {
		username, password := git.ParseUsernameAndPassword(secret.SecretContent())
		client, err = gogl.NewBasicAuthClient(nil, getBaseUrl(endpoint), username, password)
		if err != nil {
			return nil, err
		}
	}

	return &RepositoryService{
		client: client,
		repo:   repo,
	}, nil
}

func getBaseUrl(endpoint *gittransport.Endpoint) string {
	if endpoint.Protocol == "ssh" || endpoint.Protocol == "git" {
		return "https://" + endpoint.Host
	}
	return endpoint.String()[:len(endpoint.String())-len(endpoint.Path)]
}

func (s *RepositoryService) FileExistenceChecker() (repository.FileExistenceChecker, error) {
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
	return repository.NewCheckerWithFetchedFiles(filenames), nil
}

func (s *RepositoryService) GetLanguageList() ([]string, error) {
	languages, _, err := s.client.Projects.GetProjectLanguages(s.repo.OwnerWithName())
	if err != nil {
		return nil, err
	}

	return git.SortLanguagesWithFloats32(*languages), nil
}

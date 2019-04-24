package gitlab

import (
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/devconsole-git/pkg/git"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository"
	"github.com/redhat-developer/devconsole-git/pkg/log"
	gogl "github.com/xanzy/go-gitlab"
	gittransport "gopkg.in/src-d/go-git.v4/plumbing/transport"
)

const (
	gitlabHost   = "gitlab.com"
	gitlabFlavor = "gitlab"
)

type RepositoryService struct {
	clientInitializer *clientInitializer
	repo              repository.StructuredIdentifier
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
			return newGlService(gitSource, secret, endpoint)
		}
		return nil, nil
	}
}

func newGlService(gitSource *v1alpha1.GitSource, secret git.Secret, endpoint *gittransport.Endpoint) (*RepositoryService, error) {
	repo, err := repository.NewStructuredIdentifier(gitSource, endpoint)
	if err != nil {
		return nil, err
	}

	return &RepositoryService{
		clientInitializer: &clientInitializer{
			secret:   secret,
			endpoint: endpoint,
		},
		repo: repo,
	}, nil
}

type clientInitializer struct {
	client   *gogl.Client
	secret   git.Secret
	endpoint *gittransport.Endpoint
}

func (i *clientInitializer) init() (*gogl.Client, error) {
	if i.client == nil {
		client := gogl.NewClient(nil, i.secret.SecretContent())
		err := client.SetBaseURL(getBaseUrl(i.endpoint))
		if err != nil {
			return nil, err
		}

		if i.secret.SecretType() == git.UsernamePasswordType {
			username, password := git.ParseUsernameAndPassword(i.secret.SecretContent())
			client, err = gogl.NewBasicAuthClient(nil, getBaseUrl(i.endpoint), username, password)
			if err != nil {
				return nil, err
			}
		}
		i.client = client
	}
	return i.client, nil
}

func getBaseUrl(endpoint *gittransport.Endpoint) string {
	if endpoint.Protocol == "ssh" || endpoint.Protocol == "git" {
		return "https://" + endpoint.Host
	}
	return endpoint.String()[:len(endpoint.String())-len(endpoint.Path)]
}

func (s *RepositoryService) FileExistenceChecker() (repository.FileExistenceChecker, error) {
	client, err := s.clientInitializer.init()
	if err != nil {
		return nil, err
	}
	tree, _, err := client.Repositories.ListTree(
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
	client, err := s.clientInitializer.init()
	if err != nil {
		return nil, err
	}
	languages, _, err := client.Projects.GetProjectLanguages(s.repo.OwnerWithName())
	if err != nil {
		return nil, err
	}

	return git.SortLanguagesWithFloats32(*languages), nil
}
func (s *RepositoryService) CheckCredentials() error {
	client, err := s.clientInitializer.init()
	if err != nil {
		return err
	}
	_, _, err = client.Users.CurrentUser()
	return err
}

func (s *RepositoryService) CheckRepoAccessibility() error {
	client, err := s.clientInitializer.init()
	if err != nil {
		return err
	}
	_, _, err = client.Projects.GetProject(s.repo.OwnerWithName(), &gogl.GetProjectOptions{})
	return err
}

func (s *RepositoryService) CheckBranch() error {
	client, err := s.clientInitializer.init()
	if err != nil {
		return err
	}
	_, _, err = client.Branches.GetBranch(s.repo.OwnerWithName(), s.repo.Branch)
	return err
}

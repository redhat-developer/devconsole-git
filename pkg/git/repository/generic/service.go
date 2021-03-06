package generic

import (
	"fmt"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/storage"
	"strings"
	"sync"

	"github.com/redhat-developer/devconsole-git/pkg/git"
	"github.com/sirupsen/logrus"
	"gopkg.in/src-d/enry.v1"
	"gopkg.in/src-d/go-billy.v4/memfs"
	gogit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

type RepositoryService struct {
	branch               string
	repository           *gogit.Repository
	authMethod           transport.AuthMethod
	treeLoader           *treeLoader
	remoteBranchesLoader *remoteBranchesLoader
}

type treeLoader struct {
	mux  sync.Mutex
	tree *object.Tree
}

func NewRepositoryService(gitSource *v1alpha1.GitSource, secretProvider *git.SecretProvider) (repository.GitService, error) {
	store := memory.NewStorage()
	return newRepositoryService(gitSource, secretProvider.GetSecret(nil), store)
}

func newRepositoryService(gitSource *v1alpha1.GitSource, secret git.Secret, storage storage.Storer) (*RepositoryService, error) {
	branch := repository.Master
	if gitSource.Spec.Ref != "" {
		branch = gitSource.Spec.Ref
	}
	refSpec := fmt.Sprintf("+refs/heads/%[1]s:refs/remotes/origin/%[1]s", branch)
	repo, err := gogit.Init(storage, memfs.New())
	if err != nil {
		return nil, err
	}
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name:  gogit.DefaultRemoteName,
		URLs:  []string{gitSource.Spec.URL},
		Fetch: []config.RefSpec{config.RefSpec(refSpec)},
	})
	if err != nil {
		return nil, err
	}

	var authMethod transport.AuthMethod
	if secret != nil {
		authMethod, err = secret.GitAuthMethod()
		if err != nil {
			return nil, err
		}
	}

	service := &RepositoryService{
		branch:               branch,
		repository:           repo,
		authMethod:           authMethod,
		treeLoader:           &treeLoader{},
		remoteBranchesLoader: &remoteBranchesLoader{},
	}

	return service, nil
}

func (l *treeLoader) fetchTree(repository *gogit.Repository, authMethod transport.AuthMethod) (*object.Tree, error) {
	l.mux.Lock()
	defer l.mux.Unlock()

	if l.tree != nil {
		return l.tree, nil
	}
	err := repository.Fetch(&gogit.FetchOptions{
		Auth:       authMethod,
		Depth:      1,
		Tags:       gogit.NoTags,
		RemoteName: gogit.DefaultRemoteName,
	})
	if err != nil {
		return nil, err
	}
	commitIter, err := repository.CommitObjects()
	if err != nil {
		return nil, err
	}
	commitToList, err := commitIter.Next()
	if err != nil {
		return nil, err
	}

	commit, err := repository.CommitObject(commitToList.Hash)
	if err != nil {
		return nil, err
	}

	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}
	l.tree = tree
	return tree, nil
}

func (s *RepositoryService) FileExistenceChecker() (repository.FileExistenceChecker, error) {
	tree, err := s.treeLoader.fetchTree(s.repository, s.authMethod)
	if err != nil {
		return nil, err
	}
	var filenames []string
	err = tree.Files().ForEach(func(f *object.File) error {
		if !strings.Contains(f.Name, "/") {
			filenames = append(filenames, f.Name)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return repository.NewCheckerWithFetchedFiles(filenames), nil
}

func (s *RepositoryService) GetLanguageList() ([]string, error) {
	tree, err := s.treeLoader.fetchTree(s.repository, s.authMethod)
	if err != nil {
		return nil, err
	}

	languagesCounts := map[string]int{}
	err = tree.Files().ForEach(func(f *object.File) error {
		language, safe := enry.GetLanguageByExtension(f.Name)
		if safe {
			languagesCounts[language]++
		} else {
			language, safe := enry.GetLanguageByFilename(f.Name)
			if safe {
				languagesCounts[language]++
			} else {
				content, err := f.Contents()
				if err != nil {
					logrus.Warn(err)
				} else {
					language, safe = enry.GetLanguageByContent(f.Name, []byte(content))
				}
				if safe {
					languagesCounts[language]++
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return git.SortLanguagesWithInts(languagesCounts), nil
}

func (s *RepositoryService) CheckCredentials() error {
	_, err := s.remoteBranchesLoader.load(s.repository, s.authMethod)
	return err
}

func (s *RepositoryService) CheckRepoAccessibility() error {
	_, err := s.remoteBranchesLoader.load(s.repository, s.authMethod)
	return err
}

func (s *RepositoryService) CheckBranch() error {
	branches, err := s.remoteBranchesLoader.load(s.repository, s.authMethod)
	if err != nil {
		return err
	}
	for _, branch := range branches {
		if branch == s.branch {
			return nil
		}
	}
	return fmt.Errorf("branch not found")
}

type remoteBranchesLoader struct {
	remoteBranches []string
}

func (l *remoteBranchesLoader) load(repository *gogit.Repository, authMethod transport.AuthMethod) ([]string, error) {
	if l.remoteBranches != nil {
		return l.remoteBranches, nil
	}
	remote, err := repository.Remote(gogit.DefaultRemoteName)
	if err != nil {
		return nil, err
	}
	references, err := remote.List(&gogit.ListOptions{Auth: authMethod})
	if err != nil {
		return nil, err
	}
	var branches []string
	for _, ref := range references {
		branches = append(branches, ref.Name().Short())
	}
	l.remoteBranches = branches
	return branches, nil
}

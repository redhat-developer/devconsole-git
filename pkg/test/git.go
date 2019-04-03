package test

import (
	"github.com/redhat-developer/git-service/pkg/git/repository"
	"github.com/stretchr/testify/require"
	gogit "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io/ioutil"
	"testing"
	"time"
)

type DummyGitRepo struct {
	Path     string
	repo     *gogit.Repository
	workTree *gogit.Worktree
	t        *testing.T
}

func NewDummyGitRepo(t *testing.T, branchName string) *DummyGitRepo {
	path, err := ioutil.TempDir("", t.Name())
	require.NoError(t, err)
	_, err = gogit.PlainInit(path, true)
	require.NoError(t, err)

	clientPath, err := ioutil.TempDir("", t.Name()+"-client")
	require.NoError(t, err)
	client, err := gogit.PlainInit(clientPath, false)
	require.NoError(t, err)
	client.CreateRemote(&config.RemoteConfig{
		URLs: []string{path},
		Name: gogit.DefaultRemoteName,
	})

	worktree, err := client.Worktree()
	require.NoError(t, err)

	dummyGitRepo := &DummyGitRepo{
		Path:     path,
		repo:     client,
		workTree: worktree,
		t:        t,
	}
	dummyGitRepo.createInitialCommits()

	if branchName != repository.Master {
		err = worktree.Checkout(&gogit.CheckoutOptions{
			Branch: plumbing.ReferenceName("refs/heads/" + branchName),
			Create: true,
		})
		require.NoError(t, err)
	}

	return dummyGitRepo
}

func (r *DummyGitRepo) createInitialCommits() {
	r.Commit("test")

	err := r.workTree.Filesystem.Remove("test")
	require.NoError(r.t, err)
	_, err = r.workTree.Remove("test")
	require.NoError(r.t, err)
	_, err = r.workTree.Commit("removed test file", newCommitOptions())
	require.NoError(r.t, err)
	err = r.repo.Push(&gogit.PushOptions{})
	require.NoError(r.t, err)
}

func (r *DummyGitRepo) Commit(fileNames ...string) string {
	for _, fileName := range fileNames {
		_, err := r.workTree.Filesystem.Create(fileName)
		require.NoError(r.t, err)
		_, err = r.workTree.Add(fileName)
		require.NoError(r.t, err)
	}

	hash, err := r.workTree.Commit(r.t.Name(), newCommitOptions())
	require.NoError(r.t, err)
	err = r.repo.Push(&gogit.PushOptions{})
	require.NoError(r.t, err)
	return hash.String()
}

func newCommitOptions() *gogit.CommitOptions {
	return &gogit.CommitOptions{
		Author: &object.Signature{
			Name:  "anyone",
			Email: "some-email@redhat.com",
			When:  time.Now(),
		},
	}
}

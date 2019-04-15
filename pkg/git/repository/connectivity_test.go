package repository_test

import (
	"github.com/redhat-developer/devconsole-git/pkg/git/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"testing"
)

func TestIsReachableRealGHRepoWithDevelopBranch(t *testing.T) {
	// given
	endpoint, err := transport.NewEndpoint("https://github.com/fabric8-services/fabric8-tenant")
	require.NoError(t, err)

	// when
	ok, err := repository.IsReachableWithBranch(logger, "develop", endpoint)

	// then
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestIsReachableRealGLRepoWithNoBranchSet(t *testing.T) {
	// given
	endpoint, err := transport.NewEndpoint("https://gitlab.com/matousjobanek/quarkus-knative")
	require.NoError(t, err)

	// when
	ok, err := repository.IsReachableWithBranch(logger, "", endpoint)

	// then
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestIsReachableRealBBRepoWithNoBranchSet(t *testing.T) {
	// given
	endpoint, err := transport.NewEndpoint("https://bitbucket.org/mjobanek-rh/quarkus-knative")
	require.NoError(t, err)

	// when
	ok, err := repository.IsReachableWithBranch(logger, "", endpoint)

	// then
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestIsReachableRealGHNonExistingRepo(t *testing.T) {
	// given
	endpoint, err := transport.NewEndpoint("https://github.com/no-owner/no-repo")
	require.NoError(t, err)

	// when
	ok, err := repository.IsReachableWithBranch(logger, "", endpoint)

	// then
	assert.False(t, ok)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to reach the repo")
}

func TestIsReachableRealGLRepoWithNonExistingBranch(t *testing.T) {
	// given
	endpoint, err := transport.NewEndpoint("https://gitlab.com/matousjobanek/quarkus-knative")
	require.NoError(t, err)

	// when
	ok, err := repository.IsReachableWithBranch(logger, "some-cool-branch", endpoint)

	// then
	assert.False(t, ok)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to find the branch")
}

func TestIsReachableRealRepoWithDevelopBranch(t *testing.T) {
	// given
	defer gock.OffAll()
	for _, branch := range []string{"master", "dev"} {
		for _, url := range []string{
			"https://github.com/some-owner/some-repo",
			"https://github.com/some-owner/some-repo.git",
			"http://github.com/some-owner/some-repo",
			"http://github.com/some-owner/some-repo.git",
			"git@github.com:some-owner/some-repo",
			"git@github.com:some-owner/some-repo.git",
			"https://matousjobanek@github.com/some-owner/some-repo.git"} {

			endpoint, err := transport.NewEndpoint(url)
			require.NoError(t, err)
			gock.New("https://github.com").
				Get("/some-owner/some-repo.git/info/refs").
				MatchParam("service", "git-upload-pack").
				Reply(200).
				BodyString(`004a8d501bc8f3a77129c17a7120bac2d4d70f4d9291 refs/heads/master
003f8c48499a598266ed7ef609070b84d2c8707fb1dd refs/heads/dev
0000`)

			// when
			ok, err := repository.IsReachableWithBranch(logger, branch, endpoint)

			// then
			assert.True(t, ok)
			assert.NoError(t, err)
		}
	}
}

func TestIsReachableForWrongURL(t *testing.T) {
	// given
	endpoint, err := transport.NewEndpoint("some-wrong-url.com")
	require.NoError(t, err)

	// when
	ok, err := repository.IsReachableWithBranch(logger, "dev", endpoint)

	// then
	assert.False(t, ok)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to reach the repo")
}

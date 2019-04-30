package connection_test

import (
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/devconsole-git/pkg/git"
	"github.com/redhat-developer/devconsole-git/pkg/git/connection"
	"github.com/redhat-developer/devconsole-git/pkg/log"
	"github.com/redhat-developer/devconsole-git/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

var (
	logger  = &log.GitSourceLogger{Logger: logf.Log}
	homeDir = os.Getenv("HOME")
)

//
// Tests without secret
//

func TestIsReachableRealGHRepoWithDevelopBranch(t *testing.T) {
	// given
	gitSource := test.NewGitSource(
		test.WithURL("https://github.com/fabric8-services/fabric8-tenant"),
		test.WithRef("develop"))

	// when
	validationErr := connection.ValidateGitSource(logger, gitSource)

	// then
	assert.NoError(t, validationErr)
}

func TestIsReachableRealGHRepoUrlEndingWithSlash(t *testing.T) {
	// given
	gitSource := test.NewGitSource(
		test.WithURL("https://github.com/fabric8-services/fabric8-tenant/"))

	// when
	validationErr := connection.ValidateGitSource(logger, gitSource)

	// then
	assert.NoError(t, validationErr)
}

func TestIsReachableRealGLRepoWithNoBranchSet(t *testing.T) {
	// given
	gitSource := test.NewGitSource(test.WithURL("https://gitlab.com/matousjobanek/quarkus-knative"))

	// when
	validationErr := connection.ValidateGitSource(logger, gitSource)

	// then
	assert.NoError(t, validationErr)
}

func TestIsReachableRealBBRepoWithNoBranchSet(t *testing.T) {
	// given
	gitSource := test.NewGitSource(test.WithURL("https://bitbucket.org/mjobanek-rh/quarkus-knative"))

	// when
	validationErr := connection.ValidateGitSource(logger, gitSource)

	// then
	assert.NoError(t, validationErr)
}

func TestIsReachableRealGHNonExistingRepo(t *testing.T) {
	// given
	gitSource := test.NewGitSource(test.WithURL("https://github.com/no-owner/no-repo"))

	// when
	validationErr := connection.ValidateGitSource(logger, gitSource)

	// then
	require.Error(t, validationErr)
	assert.Equal(t, v1alpha1.RepoNotReachable, validationErr.Reason())
}

func TestIsReachableRealGLRepoWithNonExistingBranch(t *testing.T) {
	// given
	gitSource := test.NewGitSource(
		test.WithURL("https://gitlab.com/matousjobanek/quarkus-knative"),
		test.WithRef("some-cool-branch"))

	// when
	validationErr := connection.ValidateGitSource(logger, gitSource)

	// then
	require.Error(t, validationErr)
	assert.Equal(t, v1alpha1.BranchNotFound, validationErr.Reason())
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

			gitSource := test.NewGitSource(test.WithURL(url), test.WithRef(branch))
			gock.New("https://github.com").
				Get("/some-owner/some-repo.git/info/refs").
				MatchParam("service", "git-upload-pack").
				Reply(200).
				BodyString(`004a8d501bc8f3a77129c17a7120bac2d4d70f4d9291 refs/heads/master
003f8c48499a598266ed7ef609070b84d2c8707fb1dd refs/heads/dev
0000`)

			// when
			validationErr := connection.ValidateGitSource(logger, gitSource)

			// then
			assert.Nil(t, validationErr)
		}
	}
}

func TestIsReachableForWrongURL(t *testing.T) {
	// given
	gitSource := test.NewGitSource(test.WithURL("some-wrong-url.com"))

	// when
	validationErr := connection.ValidateGitSource(logger, gitSource)

	// then
	require.Error(t, validationErr)
	assert.Equal(t, v1alpha1.RepoNotReachable, validationErr.Reason())
}

//
// Tests with secret
//

func TestValidateRealGitHubInvalidSecret(t *testing.T) {
	// given
	glSource := test.NewGitSource(
		test.WithURL("https://github.com/MatousJobanek/quarkus-knative"))

	// when
	validationErr := connection.ValidateGitSourceWithSecret(logger, glSource, git.NewOauthToken([]byte("some-token")))

	// then
	require.Error(t, validationErr)
	assert.Equal(t, v1alpha1.BadCredentials, validationErr.Reason())
}

func TestValidateGitLabSecretAndAvailableRepo(t *testing.T) {
	// given
	defer gock.OffAll()
	glSource := test.NewGitSource(test.WithURL("https://gitlab.com/matousjobanek/quarkus-knative"))
	gock.New("https://gitlab.com/").
		Get("/api/v4/user").
		SetMatcher(test.TurnOffAllGockWhenMatched()).
		Reply(200).
		BodyString("{}")

	// when
	err := connection.ValidateGitSourceWithSecret(logger, glSource, git.NewOauthToken([]byte("")))

	// then
	require.NoError(t, err)
}

func TestValidateGitLabSecretAndAvailableRepoWithWrongBranch(t *testing.T) {
	// given
	defer gock.OffAll()
	glSource := test.NewGitSource(
		test.WithURL("https://gitlab.com/matousjobanek/quarkus-knative"),
		test.WithRef("any"))
	gock.New("https://gitlab.com/").
		Get("/api/v4/user").
		SetMatcher(test.TurnOffAllGockWhenMatched()).
		Reply(200).
		BodyString("{}")

	// when
	validationErr := connection.ValidateGitSourceWithSecret(logger, glSource, git.NewOauthToken([]byte("")))

	// then
	require.Error(t, validationErr)
	assert.Equal(t, v1alpha1.BranchNotFound, validationErr.Reason())
}

func TestValidateBitBucketWrongRepo(t *testing.T) {
	// given
	defer gock.OffAll()
	glSource := test.NewGitSource(test.WithURL("https://bitbucket.org/some-org/some-repo"))
	gock.New("https://api.bitbucket.org/").
		Get("/2.0/user").
		SetMatcher(test.TurnOffAllGockWhenMatched()).
		Reply(200)

	// when
	validationErr := connection.ValidateGitSourceWithSecret(logger, glSource, git.NewOauthToken([]byte("")))

	// then
	require.Error(t, validationErr)
	assert.Equal(t, v1alpha1.RepoNotReachable, validationErr.Reason())
}

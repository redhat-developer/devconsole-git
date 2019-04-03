package github_test

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	gogh "github.com/google/go-github/github"
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/redhat-developer/git-service/pkg/git/repository/github"
	"github.com/redhat-developer/git-service/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
	"math/rand"
	"strings"
	"testing"
)

const (
	pathToTestDir  = "../../../test"
	repoIdentifier = "some-org/some-repo"
	repoURL        = "https://github.com/" + repoIdentifier
	notFound       = `{
  "message": "Not Found",
  "documentation_url": "https://developer.github.com/v3/...."
}`
	apiRateLimit = `{
  "message": "API rate limit exceeded for 1.2.3.4. (But here's the good news: Authenticated requests get a higher rate limit. Check out the documentation for more details.)",
  "documentation_url": "https://developer.github.com/v3/#rate-limiting"
}`
)

func TestRepositoryServiceForBothAuthMethodsSuccessful(t *testing.T) {
	// given
	defer gock.OffAll()
	usernamePassword := git.NewUsernamePassword("anonymous", "")
	oauthToken := git.NewOauthToken([]byte("some-token"))

	for _, secret := range []git.Secret{usernamePassword, oauthToken} {
		mockGHCalls(t, repoIdentifier, "master", test.S("pom.xml", "mvnw"), test.S("Java", "Go"))
		source := test.NewGitSource(test.WithURL(repoURL))

		// when
		service, err := github.NewRepoServiceIfMatches()(source, secret)

		// then
		require.NoError(t, err)

		filesInRootDir, err := service.GetListOfFilesInRootDir()
		require.NoError(t, err)
		require.Len(t, filesInRootDir, 2)
		assert.Contains(t, filesInRootDir, "pom.xml")
		assert.Contains(t, filesInRootDir, "mvnw")

		languageList, err := service.GetLanguageList()
		require.NoError(t, err)
		require.Len(t, languageList, 2)
		assert.Contains(t, languageList, "Java")
		assert.Contains(t, languageList, "Go")
	}
}

func TestNewRepoServiceIfMatchesShouldNotMatchWhenSshKey(t *testing.T) {
	// given
	source := test.NewGitSource(test.WithURL("git@github.com:" + repoIdentifier))

	// when
	service, err := github.NewRepoServiceIfMatches()(source,
		git.NewSshKey(test.PrivateWithoutPassphrase(t, pathToTestDir), []byte("")))

	// then
	assert.NoError(t, err)
	assert.Nil(t, service)
}

func TestNewRepoServiceIfMatchesShouldNotMatchWhenGitLabHost(t *testing.T) {
	// given
	source := test.NewGitSource(test.WithURL("gitlab.com/" + repoIdentifier))

	// when
	service, err := github.NewRepoServiceIfMatches()(source, git.NewOauthToken([]byte("some-token")))

	// then
	assert.NoError(t, err)
	assert.Nil(t, service)
}

func TestNewRepoServiceIfMatchesShouldMatchWhenFlavorIsGitHub(t *testing.T) {
	// given
	source := test.NewGitSource(test.WithURL("gitprivatehub.com/"+repoIdentifier), test.WithFlavor("github"))

	// when
	service, err := github.NewRepoServiceIfMatches()(source, git.NewOauthToken([]byte("some-token")))

	// then
	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestNewRepoServiceIfMatchesShouldNotFailWhenSsh(t *testing.T) {
	// given
	source := test.NewGitSource(test.WithURL("git@github.com:" + repoIdentifier))

	// when
	service, err := github.NewRepoServiceIfMatches()(source, git.NewOauthToken([]byte("some-token")))

	// then
	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestRepositoryServiceForWrongRepo(t *testing.T) {
	// given
	defer gock.OffAll()
	usernamePassword := git.NewUsernamePassword("anonymous", "")
	oauthToken := git.NewOauthToken([]byte("some-token"))

	for _, secret := range []git.Secret{usernamePassword, oauthToken} {
		gock.New("https://api.github.com").
			Get(fmt.Sprintf("/repos/%s/.*", repoIdentifier)).
			Times(2).
			Reply(404).
			BodyString(notFound)
		source := test.NewGitSource(test.WithURL(repoURL), test.WithRef("dev"))

		// when
		service, err := github.NewRepoServiceIfMatches()(source, secret)

		// then
		require.NoError(t, err)

		filesInRootDir, err := service.GetListOfFilesInRootDir()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Not Found")
		require.Len(t, filesInRootDir, 0)

		languageList, err := service.GetLanguageList()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Not Found")
		require.Len(t, languageList, 0)
	}
}

func TestRepositoryServiceReturningRateLimit(t *testing.T) {
	// given
	defer gock.OffAll()
	usernamePassword := git.NewUsernamePassword("anonymous", "")
	oauthToken := git.NewOauthToken([]byte("some-token"))

	for _, secret := range []git.Secret{usernamePassword, oauthToken} {
		gock.New("https://api.github.com").
			Get(fmt.Sprintf("/repos/%s/.*", repoIdentifier)).
			Times(2).
			Reply(403).
			BodyString(apiRateLimit)
		source := test.NewGitSource(test.WithURL(repoURL))

		// when
		service, err := github.NewRepoServiceIfMatches()(source, secret)

		// then
		require.NoError(t, err)

		filesInRootDir, err := service.GetListOfFilesInRootDir()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "API rate limit exceeded")
		require.Len(t, filesInRootDir, 0)

		languageList, err := service.GetLanguageList()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "API rate limit exceeded")
		require.Len(t, languageList, 0)
	}
}

func mockGHCalls(t *testing.T, prjPath, branch string, files, langs test.SliceOfStrings) {
	var entries []gogh.TreeEntry
	for _, file := range files() {
		entries = append(entries, gogh.TreeEntry{
			SHA:  sha(file),
			Path: String(file),
		})
	}
	tree := gogh.Tree{
		SHA:       sha(files()...),
		Truncated: Boolean(false),
		Entries:   entries,
	}

	bytes, err := json.Marshal(tree)
	require.NoError(t, err)

	gock.New("https://api.github.com").
		Get(fmt.Sprintf("/repos/%s/git/trees/%s", prjPath, branch)).
		Reply(200).
		BodyString(string(bytes))

	languages := map[string]int{}
	for _, lang := range langs() {
		languages[lang] = rand.Int()
	}

	bytes, err = json.Marshal(languages)
	require.NoError(t, err)

	gock.New("https://api.github.com").
		Get(fmt.Sprintf("/repos/%s/languages", prjPath)).
		Reply(200).
		BodyString(string(bytes))
}

func sha(files ...string) *string {
	return String(string(sha1.New().Sum([]byte(strings.Join(files, "-")))))
}

func String(value string) *string {
	return &value
}
func Boolean(value bool) *bool {
	return &value
}

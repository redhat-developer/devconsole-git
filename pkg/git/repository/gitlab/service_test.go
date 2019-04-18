package gitlab_test

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/redhat-developer/devconsole-git/pkg/git"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository/gitlab"
	"github.com/redhat-developer/devconsole-git/pkg/log"
	"github.com/redhat-developer/devconsole-git/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gogl "github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
	"gopkg.in/h2non/gock.v1"
	"math/rand"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
	"testing"
)

const (
	pathToTestDir  = "../../../test"
	repoIdentifier = "some-org/some-repo"
	glHost         = "https://gitlab.com/"
	repoURL        = glHost + repoIdentifier
	notFound       = `{"message":"404 Project Not Found"}`
)

var (
	usernamePassword = git.NewUsernamePassword("anonymous", "")
	oauthToken       = git.NewOauthToken([]byte("some-token"))
	validSecrets     = []git.Secret{usernamePassword, oauthToken, nil}
	logger           = &log.GitSourceLogger{Logger: logf.Log}
)

func TestRepositoryServiceForAllValidAuthMethodsSuccessful(t *testing.T) {
	// given
	defer gock.OffAll()
	mockGLCalls(t, glHost, repoIdentifier, "master", "", test.S("pom.xml", "mvnw"), test.S("Java", "Go"))
	mockTokenCall(t)
	mockGLCalls(t, glHost, repoIdentifier, "master", "some-token", test.S("pom.xml", "mvnw"), test.S("Java", "Go"))
	mockGLCalls(t, glHost, repoIdentifier, "master", "", test.S("pom.xml", "mvnw"), test.S("Java", "Go"))

	for _, secret := range validSecrets {
		source := test.NewGitSource(test.WithURL(repoURL))

		// when
		service, err := gitlab.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))

		// then
		require.NoError(t, err)

		checker, err := service.FileExistenceChecker()
		require.NoError(t, err)
		filesInRootDir := checker.GetListOfFoundFiles()
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
	source := test.NewGitSource(test.WithURL("git@gitlab.com:" + repoIdentifier))

	// when
	service, err := gitlab.NewRepoServiceIfMatches()(logger, source,
		git.NewSecretProvider(git.NewSshKey(test.PrivateWithoutPassphrase(t, pathToTestDir), []byte(""))))

	// then
	assert.NoError(t, err)
	assert.Nil(t, service)
}

func TestNewRepoServiceIfMatchesShouldNotMatchWhenGitLabHost(t *testing.T) {
	// given
	source := test.NewGitSource(test.WithURL("gitlab.com/" + repoIdentifier))

	// when
	service, err := gitlab.NewRepoServiceIfMatches()(logger, source,
		git.NewSecretProvider(git.NewOauthToken([]byte("some-token"))))

	// then
	assert.NoError(t, err)
	assert.Nil(t, service)
}

func TestNewRepoServiceIfMatchesShouldMatchWhenFlavorIsGitHub(t *testing.T) {
	// given
	source := test.NewGitSource(test.WithURL("gitprivatelab.com/"+repoIdentifier), test.WithFlavor("gitlab"))

	// when
	service, err := gitlab.NewRepoServiceIfMatches()(logger, source,
		git.NewSecretProvider(git.NewOauthToken([]byte("some-token"))))

	// then
	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestNewRepoServiceIfMatchesShouldNotFailWhenSsh(t *testing.T) {
	// given
	source := test.NewGitSource(test.WithURL("git@gitlab.com:" + repoIdentifier))

	// when
	service, err := gitlab.NewRepoServiceIfMatches()(logger, source,
		git.NewSecretProvider(git.NewOauthToken([]byte("some-token"))))

	// then
	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestRepositoryServiceForWrongRepo(t *testing.T) {
	// given
	defer gock.OffAll()
	mockTokenCall(t)

	for _, secret := range validSecrets {
		gock.New("https://gitlab.com").
			Get(fmt.Sprintf("/api/v4/projects/%s/", repoIdentifier)).
			Times(2).
			Reply(404).
			BodyString(notFound)
		source := test.NewGitSource(test.WithURL(repoURL), test.WithRef("dev"))

		// when
		service, err := gitlab.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))

		// then
		require.NoError(t, err)

		checker, err := service.FileExistenceChecker()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Not Found")
		require.Nil(t, checker)

		languageList, err := service.GetLanguageList()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Not Found")
		require.Len(t, languageList, 0)
	}
}

func TestRepositoryServiceForPrivateInstance(t *testing.T) {
	// given
	defer gock.OffAll()
	oauthToken := git.NewOauthToken([]byte("some-token"))

	for _, url := range []string{"https://gitlab.cee.redhat.com/" + repoIdentifier, "git@gitlab.cee.redhat.com:" + repoIdentifier} {

		mockGLCalls(t, "https://gitlab.cee.redhat.com/", repoIdentifier, "master", "some-token",
			test.S("pom.xml", "mvnw"), test.S("Java", "Go"))

		source := test.NewGitSource(test.WithURL(url), test.WithFlavor("gitlab"))

		// when
		service, err := gitlab.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(oauthToken))

		// then
		require.NoError(t, err)

		checker, err := service.FileExistenceChecker()
		require.NoError(t, err)
		filesInRootDir := checker.GetListOfFoundFiles()
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

func TestRepositoryServiceCheckValidToken(t *testing.T) {
	// given
	defer gock.OffAll()

	gock.New(glHost).
		Get("/api/v4/user").
		Reply(200).
		BodyString("{}")

	source := test.NewGitSource(test.WithURL(repoURL))
	service, err := gitlab.NewRepoServiceIfMatches()(logger, source,
		git.NewSecretProvider(git.NewOauthToken([]byte("some-token"))))
	require.NoError(t, err)

	// when
	err = service.CheckCredentials()

	// then
	assert.NoError(t, err)
}

func TestRepositoryServiceCheckInvalidToken(t *testing.T) {
	// given
	defer gock.OffAll()

	gock.New(glHost).
		Get("/api/v4/user").
		Reply(401).
		BodyString("{}")

	source := test.NewGitSource(test.WithURL(repoURL))
	service, err := gitlab.NewRepoServiceIfMatches()(logger, source,
		git.NewSecretProvider(git.NewOauthToken([]byte("some-token"))))
	require.NoError(t, err)

	// when
	err = service.CheckCredentials()

	// then
	assert.Error(t, err)
}

func TestRepositoryServiceCheckValidUsernameAndPassword(t *testing.T) {
	// given
	defer gock.OffAll()

	mockTokenCall(t)
	gock.New(glHost).
		Get("/api/v4/user").
		Reply(200).
		BodyString("{}")

	source := test.NewGitSource(test.WithURL(repoURL))
	service, err := gitlab.NewRepoServiceIfMatches()(logger, source,
		git.NewSecretProvider(git.NewUsernamePassword("user", "secret")))
	require.NoError(t, err)

	// when
	err = service.CheckCredentials()

	// then
	assert.NoError(t, err)
}

func TestRepositoryServiceCheckInvalidUsernameAndPassword(t *testing.T) {
	// given
	defer gock.OffAll()

	gock.New("https://gitlab.com").
		Post("/oauth/token").
		Reply(401)

	source := test.NewGitSource(test.WithURL(repoURL))
	service, err := gitlab.NewRepoServiceIfMatches()(logger, source,
		git.NewSecretProvider(git.NewUsernamePassword("user", "secret")))
	require.NoError(t, err)

	// when
	err = service.CheckCredentials()

	// then
	assert.Error(t, err)
}

func TestRepositoryServiceCheckAccessibleRepo(t *testing.T) {
	// given
	defer gock.OffAll()

	mockTokenCall(t)
	for _, secret := range validSecrets {
		gock.New(glHost).
			Get(fmt.Sprintf("/api/v4/projects/%s", repoIdentifier)).
			Reply(200).
			BodyString("{}")

		source := test.NewGitSource(test.WithURL(repoURL))
		service, err := gitlab.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))
		require.NoError(t, err)

		// when
		err = service.CheckRepoAccessibility()

		// then
		assert.NoError(t, err)
	}
}

func TestRepositoryServiceCheckInaccessibleRepo(t *testing.T) {
	// given
	defer gock.OffAll()

	mockTokenCall(t)
	for _, secret := range validSecrets {
		gock.New(glHost).
			Get(fmt.Sprintf("/api/v4/projects/%s", repoIdentifier)).
			Reply(404)

		source := test.NewGitSource(test.WithURL(repoURL))
		service, err := gitlab.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))
		require.NoError(t, err)

		// when
		err = service.CheckRepoAccessibility()

		// then
		assert.Error(t, err)
	}
}

func TestRepositoryServiceCheckPresentBranch(t *testing.T) {
	// given
	defer gock.OffAll()

	mockTokenCall(t)
	for _, secret := range validSecrets {
		gock.New(glHost).
			Get(fmt.Sprintf("/api/v4/projects/%s/repository/branches/master", repoIdentifier)).
			Reply(200).
			BodyString("{}")

		source := test.NewGitSource(test.WithURL(repoURL))
		service, err := gitlab.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))
		require.NoError(t, err)

		// when
		err = service.CheckBranch()

		// then
		assert.NoError(t, err)
	}
}

func TestRepositoryServiceCheckMissingBranch(t *testing.T) {
	// given
	defer gock.OffAll()

	for _, secret := range validSecrets {
		gock.New(glHost).
			Get(fmt.Sprintf("/api/v4/projects/%s/repository/branches/dev", repoIdentifier)).
			Reply(404).
			BodyString("{}")

		source := test.NewGitSource(test.WithURL(repoURL), test.WithRef("dev"))
		service, err := gitlab.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))
		require.NoError(t, err)

		// when
		err = service.CheckBranch()

		// then
		assert.Error(t, err)
	}
}

func mockTokenCall(t *testing.T) {
	token := &oauth2.Token{
		AccessToken: "some-token",
		TokenType:   "bearer",
	}
	bytes, err := json.Marshal(token)
	require.NoError(t, err)
	gock.New("https://gitlab.com").
		Post("/oauth/token").
		Reply(200).
		BodyString(string(bytes))
}

func mockGLCalls(t *testing.T, host, prjPath, branch, token string, files, langs test.SliceOfStrings) {
	var treeNodes []gogl.TreeNode
	for _, file := range files() {
		treeNodes = append(treeNodes, gogl.TreeNode{
			ID:   *sha(file),
			Path: file,
			Name: file,
		})
	}

	bytes, err := json.Marshal(treeNodes)
	require.NoError(t, err)

	treeMock := gock.New(host).
		Get(fmt.Sprintf("/api/v4/projects/%s/repository/tree", prjPath)).
		MatchParam("ref", branch)
	if token != "" {
		treeMock.MatchHeader("Private-Token", token)
	}
	treeMock.Reply(200).
		BodyString(string(bytes))

	languages := map[string]float32{}
	for _, lang := range langs() {
		languages[lang] = rand.Float32()
	}

	bytes, err = json.Marshal(languages)
	require.NoError(t, err)

	langsMock := gock.New(host).
		Get(fmt.Sprintf("/api/v4/projects/%s/languages", prjPath))
	if token != "" {
		langsMock.MatchHeader("Private-Token", token)
	}
	langsMock.Reply(200).
		BodyString(string(bytes))
}

func sha(files ...string) *string {
	return String(string(sha1.New().Sum([]byte(strings.Join(files, "-")))))
}

func String(value string) *string {
	return &value
}

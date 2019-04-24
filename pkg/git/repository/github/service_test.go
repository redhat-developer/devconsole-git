package github_test

import (
	"fmt"
	"github.com/redhat-developer/devconsole-git/pkg/git"
	"github.com/redhat-developer/devconsole-git/pkg/git/detector/build"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository/github"
	"github.com/redhat-developer/devconsole-git/pkg/log"
	"github.com/redhat-developer/devconsole-git/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

const (
	ghApiHost      = "https://api.github.com"
	pathToTestDir  = "../../../test"
	repoIdentifier = "some-org/some-repo"
	repoURL        = "https://github.com/" + repoIdentifier
	apiRateLimit   = `{
  "message": "API rate limit exceeded for 1.2.3.4. (But here's the good news: Authenticated requests get a higher rate limit. Check out the documentation for more details.)",
  "documentation_url": "https://developer.github.com/v3/#rate-limiting"
}`
)

var (
	usernamePassword = git.NewUsernamePassword("some-user", "some-password")
	oauthToken       = git.NewOauthToken([]byte("some-token"))
	validSecrets     = []git.Secret{usernamePassword, oauthToken}
	logger           = &log.GitSourceLogger{Logger: logf.Log}
)

func TestRepositoryServiceForAllValidAuthMethodsSuccessful(t *testing.T) {
	// given
	defer gock.OffAll()

	for _, secret := range validSecrets {
		test.MockGHGetApiCalls(t, repoIdentifier, "master", test.S("pom.xml", "mvnw"), test.S("Java", "Go"))
		source := test.NewGitSource(test.WithURL(repoURL))

		// when
		service, err := github.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))

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
	source := test.NewGitSource(test.WithURL("git@github.com:" + repoIdentifier))

	// when
	service, err := github.NewRepoServiceIfMatches()(logger, source,
		git.NewSecretProvider(git.NewSshKey(test.PrivateWithoutPassphrase(t, pathToTestDir), []byte(""))))

	// then
	assert.NoError(t, err)
	assert.Nil(t, service)
}

func TestNewRepoServiceIfMatchesShouldNotMatchWhenGitLabHost(t *testing.T) {
	// given
	source := test.NewGitSource(test.WithURL("gitlab.com/" + repoIdentifier))

	// when
	service, err := github.NewRepoServiceIfMatches()(logger, source,
		git.NewSecretProvider(git.NewOauthToken([]byte("some-token"))))

	// then
	assert.NoError(t, err)
	assert.Nil(t, service)
}

func TestNewRepoServiceIfMatchesShouldMatchWhenFlavorIsGitHub(t *testing.T) {
	// given
	source := test.NewGitSource(test.WithURL("gitprivatehub.com/"+repoIdentifier), test.WithFlavor("github"))

	// when
	service, err := github.NewRepoServiceIfMatches()(logger, source,
		git.NewSecretProvider(git.NewOauthToken([]byte("some-token"))))

	// then
	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestNewRepoServiceIfMatchesShouldNotFailWhenSsh(t *testing.T) {
	// given
	source := test.NewGitSource(test.WithURL("git@github.com:" + repoIdentifier))

	// when
	service, err := github.NewRepoServiceIfMatches()(logger, source,
		git.NewSecretProvider(git.NewOauthToken([]byte("some-token"))))

	// then
	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestRepositoryServiceForWrongRepo(t *testing.T) {
	// given
	defer gock.OffAll()

	for _, secret := range validSecrets {
		test.MockNotFoundGitHub(repoIdentifier)
		source := test.NewGitSource(test.WithURL(repoURL), test.WithRef("dev"))

		// when
		service, err := github.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))

		// then
		require.NoError(t, err)

		checker, err := service.FileExistenceChecker()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Not Found")
		assert.Nil(t, checker)

		languageList, err := service.GetLanguageList()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Not Found")
		require.Len(t, languageList, 0)
	}
}

func TestRepositoryServiceReturningRateLimit(t *testing.T) {
	// given
	defer gock.OffAll()

	for _, secret := range validSecrets {
		gock.New("https://api.github.com").
			Get(fmt.Sprintf("/repos/%s/.*", repoIdentifier)).
			Times(2).
			Reply(403).
			BodyString(apiRateLimit)
		source := test.NewGitSource(test.WithURL(repoURL))

		// when
		service, err := github.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))

		// then
		require.NoError(t, err)

		checker, err := service.FileExistenceChecker()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "API rate limit exceeded")
		require.Nil(t, checker)

		languageList, err := service.GetLanguageList()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "API rate limit exceeded")
		require.Len(t, languageList, 0)
	}
}

func TestRepositoryServiceUsesHeadCallsWhenAnonymousSecretIsUsed(t *testing.T) {
	// given
	defer gock.OffAll()

	for _, secret := range []git.Secret{git.NewUsernamePassword("anonymous", ""), nil} {
		test.MockGHHeadCalls(repoIdentifier, "dev", test.S("pom.xml"))
		source := test.NewGitSource(test.WithURL(repoURL), test.WithRef("dev"))

		// when
		service, err := github.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))
		// then
		require.NoError(t, err)

		checker, err := service.FileExistenceChecker()
		require.NoError(t, err)
		filesInRootDir := checker.GetListOfFoundFiles()
		assert.Len(t, filesInRootDir, 0)

		// and when
		var files []string
		for _, tool := range build.Tools {
			files = append(files, checker.DetectFiles(tool)...)
		}
		assert.Len(t, files, 1)
		assert.Contains(t, files, "pom.xml")

		languageList, err := service.GetLanguageList()
		require.NoError(t, err)
		require.Len(t, languageList, 0)
	}
}

func TestRepositoryServiceCheckValidCredentials(t *testing.T) {
	// given
	defer gock.OffAll()

	for _, secret := range validSecrets {
		gock.New(ghApiHost).
			Get("user").
			Reply(200)

		source := test.NewGitSource(test.WithURL(repoURL))
		service, err := github.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))
		require.NoError(t, err)

		// when
		err = service.CheckCredentials()

		// then
		assert.NoError(t, err)
	}
}

func TestRepositoryServiceCheckInvalidCredentials(t *testing.T) {
	// given
	defer gock.OffAll()

	for _, secret := range validSecrets {
		gock.New(ghApiHost).
			Get("user").
			Reply(401)

		source := test.NewGitSource(test.WithURL(repoURL))
		service, err := github.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))
		require.NoError(t, err)

		// when
		err = service.CheckCredentials()

		// then
		assert.Error(t, err)
	}
}

func TestRepositoryServiceCheckAccessibleRepo(t *testing.T) {
	// given
	defer gock.OffAll()

	for _, secret := range validSecrets {
		gock.New(ghApiHost).
			Get(fmt.Sprintf("repos/%s", repoIdentifier)).
			Reply(200)

		source := test.NewGitSource(test.WithURL(repoURL))
		service, err := github.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))
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

	for _, secret := range validSecrets {
		gock.New(ghApiHost).
			Get(fmt.Sprintf("repos/%s", repoIdentifier)).
			Reply(404)

		source := test.NewGitSource(test.WithURL(repoURL))
		service, err := github.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))
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

	for _, secret := range validSecrets {
		gock.New(ghApiHost).
			Get(fmt.Sprintf("repos/%s/branches/master", repoIdentifier)).
			Reply(200)

		source := test.NewGitSource(test.WithURL(repoURL))
		service, err := github.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))
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
		gock.New(ghApiHost).
			Get(fmt.Sprintf("repos/%s/branches/dev", repoIdentifier)).
			Reply(404)

		source := test.NewGitSource(test.WithURL(repoURL), test.WithRef("dev"))
		service, err := github.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))
		require.NoError(t, err)

		// when
		err = service.CheckBranch()

		// then
		assert.Error(t, err)
	}
}

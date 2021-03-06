package bitbucket_test

import (
	"encoding/json"
	"fmt"
	"github.com/redhat-developer/devconsole-git/pkg/git"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository/bitbucket"
	"github.com/redhat-developer/devconsole-git/pkg/log"
	"github.com/redhat-developer/devconsole-git/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

const (
	repoIdentifier   = "some-org/some-repo"
	bbHost           = "https://bitbucket.org/"
	repoURL          = bbHost + repoIdentifier
	bbApiHost        = "https://api.bitbucket.org/"
	notFound         = `{"type":"error","error":{"message":"Repository some-non-existing-org/some-repo not found"}}`
	commitNotFound   = `{"data":{"shas":["dev"]},"type":"error","error":{"message":"Commit not found","data":{"shas":["dev"]}}}`
	tokenExpiredResp = `{"type": "error", "error": {"message": "Access token expired. Use your refresh token to obtain a new access token."}}`
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

	for _, secret := range validSecrets {
		mockBBCalls(t, bbApiHost, repoIdentifier, "master", "Java", test.S("pom.xml", "mvnw"))
		source := test.NewGitSource(test.WithURL(repoURL))

		// when
		service, err := bitbucket.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))

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
		require.Len(t, languageList, 1)
		assert.Contains(t, languageList, "Java")
	}
}

func TestNewRepoServiceIfMatchesShouldNotMatchWhenGitLabHost(t *testing.T) {
	// given
	source := test.NewGitSource(test.WithURL("gitlab.com/" + repoIdentifier))

	// when
	service, err := bitbucket.NewRepoServiceIfMatches()(logger, source,
		git.NewSecretProvider(git.NewOauthToken([]byte("some-token"))))

	// then
	assert.NoError(t, err)
	assert.Nil(t, service)
}

func TestNewRepoServiceIfMatchesShouldMatchWhenFlavorIsBitbucket(t *testing.T) {
	// given
	source := test.NewGitSource(test.WithURL("bitprivatebucket.org/"+repoIdentifier), test.WithFlavor("bitbucket"))

	// when
	service, err := bitbucket.NewRepoServiceIfMatches()(logger, source,
		git.NewSecretProvider(git.NewOauthToken([]byte("some-token"))))

	// then
	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestNewRepoServiceIfMatchesShouldNotFailWhenSsh(t *testing.T) {
	// given
	source := test.NewGitSource(test.WithURL("mjobanek@bitbucket.org:" + repoIdentifier + ".git"))

	// when
	service, err := bitbucket.NewRepoServiceIfMatches()(logger, source,
		git.NewSecretProvider(git.NewOauthToken([]byte("some-token"))))

	// then
	assert.NoError(t, err)
	assert.NotNil(t, service)
}

func TestRepositoryServiceForWrongBranch(t *testing.T) {
	// given
	defer gock.OffAll()

	for _, secret := range validSecrets {
		gock.New(bbApiHost).
			Get(fmt.Sprintf("/2.0/repositories/%s/.*", repoIdentifier)).
			Times(2).
			Reply(404).
			BodyString(commitNotFound)
		source := test.NewGitSource(test.WithURL(repoURL), test.WithRef("dev"))

		// when
		service, err := bitbucket.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))

		// then
		require.NoError(t, err)

		checker, err := service.FileExistenceChecker()
		assertErrorIsNotFound(t, err, repoIdentifier, "Commit not found")
		require.Nil(t, checker)

		languageList, err := service.GetLanguageList()
		assertErrorIsNotFound(t, err, repoIdentifier, "Commit not found")
		require.Len(t, languageList, 0)
	}
}

func TestRepositoryServiceForWrongRepo(t *testing.T) {
	// given
	defer gock.OffAll()

	for _, secret := range validSecrets {
		gock.New(bbApiHost).
			Get("/2.0/repositories/some-non-existing-org/some-repo/.*").
			Times(2).
			Reply(404).
			BodyString(notFound)
		source := test.NewGitSource(test.WithURL(bbHost + "some-non-existing-org/some-repo"))

		// when
		service, err := bitbucket.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))

		// then
		require.NoError(t, err)

		checker, err := service.FileExistenceChecker()
		assertErrorIsNotFound(t, err, "some-non-existing-org/some-repo", "Repository some-non-existing-org/some-repo not found")
		require.Nil(t, checker)

		languageList, err := service.GetLanguageList()
		assertErrorIsNotFound(t, err, "some-non-existing-org/some-repo", "Repository some-non-existing-org/some-repo not found")
		require.Len(t, languageList, 0)
	}
}

func assertErrorIsNotFound(t *testing.T, err error, repo, msg string) {
	require.Error(t, err)
	assert.Contains(t, err.Error(), "call to the API endpoint https://api.bitbucket.org/2.0/repositories/"+repo)
	assert.Contains(t, err.Error(), fmt.Sprintf("failed with [404 Not Found] and message [%s]", msg))
}

func TestRepositoryServiceReturningForbidden(t *testing.T) {
	// given
	defer gock.OffAll()

	for _, secret := range validSecrets {
		gock.New(bbApiHost).
			Get(fmt.Sprintf("/2.0/repositories/%s/.*", repoIdentifier)).
			Times(2).
			Reply(403).
			BodyString("Forbidden")
		source := test.NewGitSource(test.WithURL(repoURL))

		// when
		service, err := bitbucket.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))

		// then
		require.NoError(t, err)

		checker, err := service.FileExistenceChecker()
		assertErrorIsForbidden(t, err)
		require.Nil(t, checker)

		languageList, err := service.GetLanguageList()
		assertErrorIsForbidden(t, err)
		require.Len(t, languageList, 0)
	}
}

func assertErrorIsForbidden(t *testing.T, err error) {
	require.Error(t, err)
	assert.Contains(t, err.Error(), "call to the API endpoint https://api.bitbucket.org/2.0/repositories/"+repoIdentifier)
	assert.Contains(t, err.Error(), "failed with [403 Forbidden] and message [Forbidden]")
}

func TestRepositoryServiceReturningTokenExpired(t *testing.T) {
	// given
	defer gock.OffAll()
	oauthToken := git.NewOauthToken([]byte("some-token"))

	gock.New(bbApiHost).
		Get(fmt.Sprintf("/2.0/repositories/%s/.*", repoIdentifier)).
		Times(2).
		Reply(401).
		BodyString(tokenExpiredResp)
	source := test.NewGitSource(test.WithURL(repoURL))

	// when
	service, err := bitbucket.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(oauthToken))

	// then
	require.NoError(t, err)

	checker, err := service.FileExistenceChecker()
	assertErrorIsTokenExp(t, err)
	require.Nil(t, checker)

	languageList, err := service.GetLanguageList()
	assertErrorIsTokenExp(t, err)
	require.Len(t, languageList, 0)
}

func assertErrorIsTokenExp(t *testing.T, err error) {
	require.Error(t, err)
	assert.Contains(t, err.Error(), "call to the API endpoint https://api.bitbucket.org/2.0/repositories/"+repoIdentifier)
	assert.Contains(t, err.Error(), "failed with [401 Unauthorized] and message [Access token expired. Use your refresh token to obtain a new access token.]")
}

func TestRepositoryServiceForPrivateInstance(t *testing.T) {
	// given
	defer gock.OffAll()
	oauthToken := git.NewOauthToken([]byte("some-token"))

	for _, url := range []string{
		"https://bitbucket.redhat.com/" + repoIdentifier, "git@bitbucket.redhat.com:" + repoIdentifier + ".git"} {
		mockBBCalls(t, "https://api.bitbucket.redhat.com/", repoIdentifier, "master", "Java", test.S("pom.xml", "mvnw"))

		source := test.NewGitSource(test.WithURL(url), test.WithFlavor("bitbucket"))

		// when
		service, err := bitbucket.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(oauthToken))

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
		require.Len(t, languageList, 1)
		assert.Contains(t, languageList, "Java")
	}
}

func TestRepositoryServiceWithPaginatedResult(t *testing.T) {
	// given
	defer gock.OffAll()
	baseURL := fmt.Sprintf(`%s2.0/repositories/%s/src/master/?q=type="commit_file"`, bbApiHost, repoIdentifier)

	for _, secret := range validSecrets {
		mockBBFilesCall(t, bbApiHost, repoIdentifier, "master", "", baseURL+"&page=8xhd", test.S("pom.xml"))
		mockBBFilesCall(t, bbApiHost, repoIdentifier, "master", "8xhd", baseURL+"&page=dbtR", test.S("mvnw"))
		mockBBFilesCall(t, bbApiHost, repoIdentifier, "master", "dbtR", "", test.S("any"))
		mockBBRepoCall(t, bbApiHost, repoIdentifier, "Java")

		source := test.NewGitSource(test.WithURL(repoURL))

		// when
		service, err := bitbucket.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))

		// then
		require.NoError(t, err)

		checker, err := service.FileExistenceChecker()
		require.NoError(t, err)
		filesInRootDir := checker.GetListOfFoundFiles()
		require.Len(t, filesInRootDir, 3)
		assert.Contains(t, filesInRootDir, "pom.xml")
		assert.Contains(t, filesInRootDir, "mvnw")
		assert.Contains(t, filesInRootDir, "any")

		languageList, err := service.GetLanguageList()
		require.NoError(t, err)
		require.Len(t, languageList, 1)
		assert.Contains(t, languageList, "Java")
	}
}

func TestRepositoryServiceCheckValidCredentials(t *testing.T) {
	// given
	defer gock.OffAll()

	for _, secret := range validSecrets {
		gock.New(bbApiHost).
			Get("/2.0/user").
			Reply(200)

		source := test.NewGitSource(test.WithURL(repoURL))
		service, err := bitbucket.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))
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
		gock.New(bbApiHost).
			Get("/2.0/user").
			Reply(401)

		source := test.NewGitSource(test.WithURL(repoURL))
		service, err := bitbucket.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))
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
		mockBBRepoCall(t, bbApiHost, repoIdentifier, "")

		source := test.NewGitSource(test.WithURL(repoURL))
		service, err := bitbucket.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))
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

	for _, reply := range []int{404, 401} {
		for _, secret := range validSecrets {
			gock.New(bbApiHost).
				Get(fmt.Sprintf("/2.0/repositories/%s/", repoIdentifier)).
				Reply(reply)

			source := test.NewGitSource(test.WithURL(repoURL))
			service, err := bitbucket.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))
			require.NoError(t, err)

			// when
			err = service.CheckRepoAccessibility()

			// then
			assert.Error(t, err)
		}
	}
}

func TestRepositoryServiceCheckPresentBranch(t *testing.T) {
	// given
	defer gock.OffAll()

	for _, secret := range validSecrets {
		gock.New(bbApiHost).
			Get(fmt.Sprintf("/2.0/repositories/%s/refs/branches/master", repoIdentifier)).
			Reply(200)

		source := test.NewGitSource(test.WithURL(repoURL))
		service, err := bitbucket.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))
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
		gock.New(bbApiHost).
			Get(fmt.Sprintf("/2.0/repositories/%s/refs/branches/dev", repoIdentifier)).
			Reply(404)

		source := test.NewGitSource(test.WithURL(repoURL), test.WithRef("dev"))
		service, err := bitbucket.NewRepoServiceIfMatches()(logger, source, git.NewSecretProvider(secret))
		require.NoError(t, err)

		// when
		err = service.CheckBranch()

		// then
		assert.Error(t, err)
	}
}

func mockBBCalls(t *testing.T, host, prjPath, branch, lang string, files test.SliceOfStrings) {
	mockBBFilesCall(t, host, prjPath, branch, "", "", files)
	mockBBRepoCall(t, host, prjPath, lang)
}

func mockBBRepoCall(t *testing.T, host, prjPath, lang string) {
	repoLang := bitbucket.RepositoryLanguage{Language: lang}

	bytes, err := json.Marshal(repoLang)
	require.NoError(t, err)

	gock.New(host).
		Get(fmt.Sprintf("/2.0/repositories/%s/", prjPath)).
		Reply(200).
		BodyString(string(bytes))
}

func mockBBFilesCall(t *testing.T, host, prjPath, branch, pageParam, nextPage string, files test.SliceOfStrings) {
	var entries []bitbucket.FileEntry
	for _, file := range files() {
		entries = append(entries, bitbucket.FileEntry{
			Path: file,
		})
	}
	src := bitbucket.Src{
		Pagination: bitbucket.Pagination{
			Page: 1,
			Next: nextPage,
		},
		Values: entries,
	}

	bytes, err := json.Marshal(src)
	require.NoError(t, err)

	mock := gock.New(host).
		Get(fmt.Sprintf(`/2.0/repositories/%s/src/%s/`, prjPath, branch)).
		MatchParam("q", `type="commit_file"`)
	if pageParam != "" {
		mock.MatchParam("page", pageParam)
	}
	mock.Reply(200).
		BodyString(string(bytes))
}

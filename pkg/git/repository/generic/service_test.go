package generic_test

import (
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/redhat-developer/git-service/pkg/git/repository"
	"github.com/redhat-developer/git-service/pkg/git/repository/generic"
	"github.com/redhat-developer/git-service/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

const pathToTestDir = "../../../test"

func TestNewRepositoryServiceForAllSecretsAndMethods(t *testing.T) {
	// given
	dummyRepo := test.NewDummyGitRepo(t, repository.Master)
	dummyRepo.Commit("pom.xml", "mvnw", "src/main/java/Any.java", "pkg/main.go")
	sshKey := git.NewSshKey(test.PrivateWithoutPassphrase(t, pathToTestDir), []byte(""))
	usernamePassword := git.NewUsernamePassword("anonymous", "")
	oauthToken := git.NewOauthToken([]byte("some-token"))

	for _, secret := range []git.Secret{sshKey, usernamePassword, oauthToken, nil} {
		source := test.NewGitSource(test.WithURL(dummyRepo.Path))

		// when
		service, err := generic.NewRepositoryService(source, git.NewSecretProvider(secret))

		// then
		require.NoError(t, err)
		rootFiles, err := service.GetListOfFilesInRootDir()
		require.NoError(t, err)
		require.Len(t, rootFiles, 2)
		assert.Contains(t, rootFiles, "pom.xml")
		assert.Contains(t, rootFiles, "mvnw")

		languageList, err := service.GetLanguageList()
		require.NoError(t, err)
		assert.Len(t, languageList, 3)
		assert.Contains(t, languageList, "XML")
		assert.Contains(t, languageList, "Java")
		assert.Contains(t, languageList, "Go")
	}
}

func TestNewRepositoryServiceShouldReturnFilesAddedByMultipleCommits(t *testing.T) {
	// given
	dummyRepo := test.NewDummyGitRepo(t, repository.Master)
	dummyRepo.Commit("pom.xml")
	dummyRepo.Commit("mvnw")
	dummyRepo.Commit("main.go")
	source := test.NewGitSource(test.WithURL(dummyRepo.Path))

	// when
	service, err := generic.NewRepositoryService(source,
		git.NewSecretProvider(git.NewSshKey(test.PrivateWithoutPassphrase(t, pathToTestDir), []byte(""))))

	// then
	require.NoError(t, err)
	rootFiles, err := service.GetListOfFilesInRootDir()
	require.NoError(t, err)
	require.Len(t, rootFiles, 3)
	assert.Contains(t, rootFiles, "pom.xml")
	assert.Contains(t, rootFiles, "mvnw")
	assert.Contains(t, rootFiles, "main.go")
}

func TestNewRepositoryServiceWithEmptyMasterBranch(t *testing.T) {
	// given
	dummyRepo := test.NewDummyGitRepo(t, "dev")
	dummyRepo.Commit("pom.xml", "mvnw", "src/main/java/Any.java", "pkg/main.go")
	source := test.NewGitSource(test.WithURL(dummyRepo.Path))

	// when
	service, err := generic.NewRepositoryService(source,
		git.NewSecretProvider(git.NewSshKey(test.PrivateWithoutPassphrase(t, pathToTestDir), []byte(""))))

	// then
	require.NoError(t, err)
	rootFiles, err := service.GetListOfFilesInRootDir()
	require.NoError(t, err)
	require.Len(t, rootFiles, 0)

	languageList, err := service.GetLanguageList()
	require.NoError(t, err)
	assert.Len(t, languageList, 0)
}

func TestNewRepositoryServiceWithOtherThanMasterBranch(t *testing.T) {
	// given
	dummyRepo := test.NewDummyGitRepo(t, "dev")
	dummyRepo.Commit("main.go", "any-file")
	source := test.NewGitSource(test.WithURL(dummyRepo.Path), test.WithRef("dev"))

	// when
	service, err := generic.NewRepositoryService(source,
		git.NewSecretProvider(git.NewSshKey(test.PrivateWithoutPassphrase(t, pathToTestDir), []byte(""))))

	// then
	require.NoError(t, err)
	rootFiles, err := service.GetListOfFilesInRootDir()
	require.NoError(t, err)
	require.Len(t, rootFiles, 2)
	assert.Contains(t, rootFiles, "main.go")
	assert.Contains(t, rootFiles, "any-file")

	languageList, err := service.GetLanguageList()
	require.NoError(t, err)
	assert.Len(t, languageList, 1)
	assert.Contains(t, languageList, "Go")
}

func TestNewRepositoryServiceShouldLoadFilesOnlyOnce(t *testing.T) {
	// given
	dummyRepo := test.NewDummyGitRepo(t, repository.Master)
	dummyRepo.Commit("main.go")
	source := test.NewGitSource(test.WithURL(dummyRepo.Path))

	service, err := generic.NewRepositoryService(source,
		git.NewSecretProvider(git.NewSshKey(test.PrivateWithoutPassphrase(t, pathToTestDir), []byte(""))))
	require.NoError(t, err)

	// when
	rootFiles, err := service.GetListOfFilesInRootDir()

	// then
	require.NoError(t, err)
	require.Len(t, rootFiles, 1)
	assert.Contains(t, rootFiles, "main.go")

	// and when
	dummyRepo.Commit("pom.xml")
	rootFiles, err = service.GetListOfFilesInRootDir()

	// then it should use cache
	require.Len(t, rootFiles, 1)
	assert.Contains(t, rootFiles, "main.go")
	languageList, err := service.GetLanguageList()
	require.NoError(t, err)
	assert.Len(t, languageList, 1)
	assert.Contains(t, languageList, "Go")
}

func TestNewRepositoryServiceUsingSSh(t *testing.T) {
	// given
	allowedPubKey := test.PublicWithoutPassphrase(t, pathToTestDir)
	reset := test.RunKeySshServer(t, allowedPubKey)
	defer reset()

	dummyRepo := test.NewDummyGitRepo(t, repository.Master)
	dummyRepo.Commit("main.go")
	source := test.NewGitSource(test.WithURL("ssh://git@localhost:2222" + dummyRepo.Path))

	service, err := generic.NewRepositoryService(source,
		git.NewSecretProvider(git.NewSshKey(test.PrivateWithoutPassphrase(t, pathToTestDir), []byte(""))))
	require.NoError(t, err)

	// when
	rootFiles, err := service.GetListOfFilesInRootDir()

	// then
	require.NoError(t, err)
	require.Len(t, rootFiles, 1)
	assert.Contains(t, rootFiles, "main.go")
	languageList, err := service.GetLanguageList()
	require.NoError(t, err)
	assert.Len(t, languageList, 1)
	assert.Contains(t, languageList, "Go")
}

func TestNewRepositoryServiceUsingSShWithWrongKey(t *testing.T) {
	// given
	allowedPubKey := test.PublicWithoutPassphrase(t, pathToTestDir)
	reset := test.RunKeySshServer(t, allowedPubKey)
	defer reset()

	dummyRepo := test.NewDummyGitRepo(t, repository.Master)
	dummyRepo.Commit("main.go")
	source := test.NewGitSource(test.WithURL("ssh://git@localhost:2222" + dummyRepo.Path))

	service, err := generic.NewRepositoryService(source,
		git.NewSecretProvider(git.NewSshKey(test.PrivateWithPassphrase(t, pathToTestDir), []byte("secret"))))
	require.NoError(t, err)

	// when
	_, err = service.GetListOfFilesInRootDir()

	// then
	assert.Contains(t, err.Error(), "ssh: handshake failed: ssh: unable to authenticate")

	// and when
	_, err = service.GetLanguageList()

	// then
	assert.Contains(t, err.Error(), "ssh: handshake failed: ssh: unable to authenticate")
}

func TestNewRepositoryServiceUsingBasicSShAuth(t *testing.T) {
	// given
	reset := test.RunBasicSshServer(t, "super-secret")
	defer reset()

	dummyRepo := test.NewDummyGitRepo(t, repository.Master)
	dummyRepo.Commit("main.go")
	usernamePassword := git.NewUsernamePassword("user", "super-secret")
	oauthToken := git.NewOauthToken([]byte("super-secret"))
	for _, secret := range []git.Secret{usernamePassword, oauthToken} {
		source := test.NewGitSource(test.WithURL("ssh://git@localhost:2222" + dummyRepo.Path))

		service, err := generic.NewRepositoryService(source, git.NewSecretProvider(secret))
		require.NoError(t, err)

		// when
		rootFiles, err := service.GetListOfFilesInRootDir()

		// then
		require.NoError(t, err)
		require.Len(t, rootFiles, 1)
		assert.Contains(t, rootFiles, "main.go")
		languageList, err := service.GetLanguageList()
		require.NoError(t, err)
		assert.Len(t, languageList, 1)
		assert.Contains(t, languageList, "Go")
	}
}

func TestNewRepositoryServiceUsingBasicSShAuthWithWrongPassword(t *testing.T) {
	// given
	reset := test.RunBasicSshServer(t, "super-secret")
	defer reset()

	dummyRepo := test.NewDummyGitRepo(t, repository.Master)
	dummyRepo.Commit("main.go")
	usernamePassword := git.NewUsernamePassword("user", "wrong-secret")
	oauthToken := git.NewOauthToken([]byte("wrong-secret"))
	for _, secret := range []git.Secret{usernamePassword, oauthToken} {
		source := test.NewGitSource(test.WithURL("ssh://git@localhost:2222" + dummyRepo.Path))

		service, err := generic.NewRepositoryService(source, git.NewSecretProvider(secret))
		require.NoError(t, err)

		// when
		_, err = service.GetListOfFilesInRootDir()

		// then
		assert.Contains(t, err.Error(), "ssh: handshake failed: ssh: unable to authenticate")

		// and when
		_, err = service.GetLanguageList()

		// then
		assert.Contains(t, err.Error(), "ssh: handshake failed: ssh: unable to authenticate")
	}
}

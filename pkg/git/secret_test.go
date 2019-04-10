package git_test

import (
	"bytes"
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/redhat-developer/git-service/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	gogitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"testing"
)

var pathToTestDir = "../test"

func TestNewSshKey(t *testing.T) {
	// given
	privateKey := test.PrivateWithPassphrase(t, pathToTestDir)

	// when
	key := git.NewSshKey(privateKey, []byte("secret"))

	// then
	assert.Equal(t, string(bytes.TrimSpace(privateKey)), key.SecretContent())
	assert.Equal(t, git.SshKeyType, key.SecretType())
	assert.Nil(t, key.Client())
}

func TestNewSshKeyWithPassphraseAuthMethodSuccessful(t *testing.T) {
	// given
	key := git.NewSshKey(test.PrivateWithPassphrase(t, pathToTestDir), []byte("secret"))

	// when
	authMethod, err := key.GitAuthMethod()

	// then
	require.NoError(t, err)
	assertSshKey(t, authMethod, test.PublicWithPassphrase(t, pathToTestDir))
}

func TestNewSshKeyWithoutPassphraseAuthMethodUsingEmptyPassphraseSuccessful(t *testing.T) {
	// given
	key := git.NewSshKey(test.PrivateWithoutPassphrase(t, pathToTestDir), []byte(""))

	// when
	authMethod, err := key.GitAuthMethod()

	// then
	require.NoError(t, err)
	assertSshKey(t, authMethod, test.PublicWithoutPassphrase(t, pathToTestDir))
}

func TestNewSshKeyWithPassphraseAuthMethodUsingEmptyPassphraseFailed(t *testing.T) {
	// given
	key := git.NewSshKey(test.PrivateWithPassphrase(t, pathToTestDir), []byte(""))

	// when
	authMethod, err := key.GitAuthMethod()

	// then
	assert.Error(t, err)
	assert.Nil(t, authMethod)
}

func TestNewSshKeyWithPassphraseAuthMethodUsingAnyPassphraseFailed(t *testing.T) {
	// given
	key := git.NewSshKey(test.PrivateWithPassphrase(t, pathToTestDir), []byte("any"))

	// when
	authMethod, err := key.GitAuthMethod()

	// then
	assert.Error(t, err)
	assert.Nil(t, authMethod)
}

func TestNewSshKeyWithoutPassphraseAuthMethodUsingAnyPassphraseSuccessful(t *testing.T) {
	// given
	key := git.NewSshKey(test.PrivateWithoutPassphrase(t, pathToTestDir), []byte("any"))

	// when
	authMethod, err := key.GitAuthMethod()

	// then
	require.NoError(t, err)
	assertSshKey(t, authMethod, test.PublicWithoutPassphrase(t, pathToTestDir))
}

func assertSshKey(t *testing.T, authMethod transport.AuthMethod, publicKeyContent []byte) {
	assert.Equal(t, gogitssh.PublicKeysName, authMethod.Name())
	keys, ok := authMethod.(*gitssh.PublicKeys)
	require.True(t, ok)

	publicKey, _, _, _, err := ssh.ParseAuthorizedKey(publicKeyContent)
	require.NoError(t, err)
	assert.Equal(t, publicKey.Marshal(), keys.Signer.PublicKey().Marshal())
	assert.Equal(t, "git", keys.User)
}

func TestNewOauthToken(t *testing.T) {
	// given
	token := "some-token"

	// when
	oauthToken := git.NewOauthToken([]byte(token))

	// then
	assert.Equal(t, token, oauthToken.SecretContent())
	assert.Equal(t, git.OauthTokenType, oauthToken.SecretType())

	client := oauthToken.Client()
	require.NotNil(t, client)
	transport, ok := client.Transport.(*oauth2.Transport)
	assert.True(t, ok)

	tokenSource, err := transport.Source.Token()
	require.NoError(t, err)
	assert.Equal(t, "", tokenSource.TokenType)
	assert.Equal(t, token, tokenSource.AccessToken)

	method, err := oauthToken.GitAuthMethod()
	require.NoError(t, err)
	assertBasicAuth(t, method, "", token)
}

func assertBasicAuth(t *testing.T, authMethod transport.AuthMethod, username, password string) {
	assert.Equal(t, gogitssh.PasswordName, authMethod.Name())
	basic, ok := authMethod.(*gogitssh.Password)
	require.True(t, ok)

	assert.Equal(t, username, basic.User)
	assert.Equal(t, password, basic.Password)
}

func TestNewUsernamePassword(t *testing.T) {
	// when
	basic := git.NewUsernamePassword("username", "password")

	// then
	assert.Equal(t, git.UsernamePasswordType, basic.SecretType())
	assert.Equal(t, "username:password", basic.SecretContent())

	client := basic.Client()
	require.NotNil(t, client)

	method, err := basic.GitAuthMethod()
	require.NoError(t, err)
	assertBasicAuth(t, method, "username", "password")
}

func TestNewSecretProviderWithSecretSet(t *testing.T) {
	// given
	oauthToken := git.NewOauthToken([]byte("some-token"))
	secretProvider := git.NewSecretProvider(oauthToken)

	// when
	secret := secretProvider.GetSecret(git.NewUsernamePassword("anonymous", ""))

	// then
	assert.Equal(t, oauthToken, secret)
}

func TestNewSecretProviderWithSecretNotSet(t *testing.T) {
	// given
	secretProvider := git.NewSecretProvider(nil)
	usernamePassword := git.NewUsernamePassword("anonymous", "")

	// when
	secret := secretProvider.GetSecret(usernamePassword)

	// then
	assert.Equal(t, usernamePassword, secret)
}

package git_test

import (
	"bytes"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/redhat-developer/git-service/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
	"gopkg.in/h2non/gock.v1"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"

	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	gogitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

const pathToTestDir = "../test"

var defaultToken = git.NewOauthToken([]byte("default"))

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

func TestGetGitSecretWithNoCredentials(t *testing.T) {
	//given
	gs := test.NewGitSource()
	client, _ := test.PrepareClient(test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs))

	//when
	secretProvider, err := git.NewGitSecretProvider(client, test.Namespace, gs)

	//then
	require.NoError(t, err)
	assert.Empty(t, secretProvider.SecretType())
	assert.Equal(t, defaultToken, secretProvider.GetSecret(defaultToken))
}

func TestGetGitSecretWithBasicCredentials(t *testing.T) {
	//given
	for _, secretType := range []corev1.SecretType{corev1.SecretTypeOpaque, corev1.SecretTypeBasicAuth} {
		sec := test.NewSecret(secretType, map[string][]byte{
			"username": []byte("username"),
			"password": []byte("password")})

		gs := test.NewGitSource()
		gs.Spec.SecretRef = &v1alpha1.SecretRef{Name: test.SecretName}
		client, _ := test.PrepareClient(
			test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs),
			test.RegisterGvkObject(corev1.SchemeGroupVersion, sec))

		//when
		secretProvider, err := git.NewGitSecretProvider(client, test.Namespace, gs)

		//then
		require.NoError(t, err)
		secret := secretProvider.GetSecret(defaultToken)
		assert.Equal(t, git.UsernamePasswordType, secret.SecretType())
		assert.Equal(t, "username:password", secret.SecretContent())
	}
}

func TestGetGitSecretWithWrongSecret(t *testing.T) {
	//given
	defer gock.OffAll()
	sec := test.NewSecret(corev1.SecretTypeTLS, map[string][]byte{"tls.crt": []byte("crt")})

	gs := test.NewGitSource()
	gs.Spec.SecretRef = &v1alpha1.SecretRef{Name: test.SecretName}
	client, _ := test.PrepareClient(
		test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs),
		test.RegisterGvkObject(corev1.SchemeGroupVersion, sec))

	//when
	secretProvider, err := git.NewGitSecretProvider(client, test.Namespace, gs)

	//then
	assert.Error(t, err)
	assert.Nil(t, secretProvider)
}

func TestGetGitSecretWithGivenToken(t *testing.T) {
	//given
	defer gock.OffAll()
	for _, secretType := range []corev1.SecretType{corev1.SecretTypeOpaque, corev1.SecretTypeBasicAuth} {
		sec := test.NewSecret(secretType, map[string][]byte{"password": []byte("some-token")})

		gs := test.NewGitSource()
		gs.Spec.SecretRef = &v1alpha1.SecretRef{Name: test.SecretName}
		client, _ := test.PrepareClient(
			test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs),
			test.RegisterGvkObject(corev1.SchemeGroupVersion, sec))

		//when
		secretProvider, err := git.NewGitSecretProvider(client, test.Namespace, gs)

		//then
		require.NoError(t, err)
		secret := secretProvider.GetSecret(defaultToken)
		assert.Equal(t, git.OauthTokenType, secret.SecretType())
		assert.Equal(t, "some-token", secret.SecretContent())
	}
}

func TestGetGitSecretWithSshKey(t *testing.T) {
	//given
	for _, secretType := range []corev1.SecretType{corev1.SecretTypeOpaque, corev1.SecretTypeSSHAuth} {
		sec := test.NewSecret(secretType, map[string][]byte{
			"ssh-privatekey": test.PrivateWithoutPassphrase(t, pathToTestDir)})

		gs := test.NewGitSource()
		gs.Spec.SecretRef = &v1alpha1.SecretRef{Name: test.SecretName}
		client, _ := test.PrepareClient(
			test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs),
			test.RegisterGvkObject(corev1.SchemeGroupVersion, sec))

		//when
		secretProvider, err := git.NewGitSecretProvider(client, test.Namespace, gs)

		//then
		require.NoError(t, err)
		secret := secretProvider.GetSecret(defaultToken)
		assert.Equal(t, git.SshKeyType, secret.SecretType())
		assert.NotEmpty(t, secret.SecretContent())
	}
}

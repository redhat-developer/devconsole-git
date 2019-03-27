package git_test

import (
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	gogitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"testing"
)

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

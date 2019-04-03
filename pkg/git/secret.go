package git

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

type Secret interface {
	// GitAuthMethod returns an instance of git AuthMethod for the secret
	GitAuthMethod() (transport.AuthMethod, error)
	// Client returns an instance of http.Client for the secret
	Client() *http.Client
	// SecretType returns a type of the secret
	SecretType() string
	// SecretContent returns an actual content of the secret
	SecretContent() string
}

type commonSecretInfo struct {
	secretType    string
	secretContent []byte
}

func (k *commonSecretInfo) SecretType() string {
	return k.secretType
}

func (k *commonSecretInfo) SecretContent() string {
	return string(k.secretContent)
}

const (
	OauthTokenType       = "OauthToken"
	UsernamePasswordType = "UsernamePassword"
)

type SshKey struct {
	*commonSecretInfo
	passphrase []byte
}

var allowAll = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
	return nil
}

type OauthToken struct {
	*commonSecretInfo
}

func NewOauthToken(token []byte) *OauthToken {
	return &OauthToken{
		commonSecretInfo: &commonSecretInfo{
			secretType:    OauthTokenType,
			secretContent: bytes.TrimSpace(token),
		},
	}
}

func (t *OauthToken) GitAuthMethod() (transport.AuthMethod, error) {
	return &gitssh.Password{
		Password: string(t.secretContent),
		HostKeyCallbackHelper: gitssh.HostKeyCallbackHelper{
			HostKeyCallback: allowAll,
		}}, nil
}

func (t *OauthToken) Client() *http.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: string(t.secretContent)},
	)
	return oauth2.NewClient(ctx, ts)

}

type UsernamePassword struct {
	*commonSecretInfo
	username string
	password string
}

func NewUsernamePassword(username, password string) *UsernamePassword {
	return &UsernamePassword{
		commonSecretInfo: &commonSecretInfo{
			secretType:    UsernamePasswordType,
			secretContent: []byte(fmt.Sprintf("%s:%s", username, password)),
		},
		username: username,
		password: password,
	}
}

func (t *UsernamePassword) GitAuthMethod() (transport.AuthMethod, error) {
	return &gitssh.Password{
		User:     t.username,
		Password: t.password,
		HostKeyCallbackHelper: gitssh.HostKeyCallbackHelper{
			HostKeyCallback: allowAll,
		}}, nil
}

func (t *UsernamePassword) Client() *http.Client {
	return &http.Client{}
}

func ParseUsernameAndPassword(secret string) (string, string) {
	split := strings.Split(secret, ":")
	if len(split) == 2 {
		return split[0], split[1]
	}
	return "", ""
}

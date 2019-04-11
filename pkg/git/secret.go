package git

import (
	"bytes"
	"context"
	"fmt"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"net"
	"net/http"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SecretProvider struct {
	secret Secret
}

func NewSecretProvider(secret Secret) *SecretProvider {
	return &SecretProvider{secret: secret}
}

func (p *SecretProvider) GetSecret(defaultSecret Secret) Secret {
	if p.secret == nil {
		return defaultSecret
	}
	return p.secret
}

func (p *SecretProvider) SecretType() string {
	if p.secret == nil {
		return ""
	}
	return p.secret.SecretType()
}

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
	SshKeyType           = "SshKey"
	OauthTokenType       = "OauthToken"
	UsernamePasswordType = "UsernamePassword"
)

type SshKey struct {
	*commonSecretInfo
	passphrase []byte
}

func NewSshKey(sshKey []byte, passphrase []byte) *SshKey {
	return &SshKey{
		commonSecretInfo: &commonSecretInfo{
			secretType:    SshKeyType,
			secretContent: bytes.TrimSpace(sshKey),
		},
		passphrase: passphrase,
	}
}

var allowAll = func(hostname string, remote net.Addr, key ssh.PublicKey) error {
	return nil
}

func (k *SshKey) GitAuthMethod() (transport.AuthMethod, error) {
	var signer ssh.Signer
	var err error
	if len(k.passphrase) > 0 {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(k.secretContent, k.passphrase)
		if err != nil {
			return nil, err
		}
	} else {
		signer, err = ssh.ParsePrivateKey(k.secretContent)
		if err != nil {
			return nil, err
		}
	}

	return &gitssh.PublicKeys{
		User:   "git",
		Signer: signer,
		HostKeyCallbackHelper: gitssh.HostKeyCallbackHelper{
			HostKeyCallback: allowAll,
		},
	}, nil
}

func (k *SshKey) Client() *http.Client {
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

func NewGitSecretProvider(client client.Client, namespace string, gitSource *v1alpha1.GitSource) (*SecretProvider, error) {
	if gitSource.Spec.SecretRef == nil {
		return NewSecretProvider(nil), nil
	}
	coreSecret := &corev1.Secret{}
	namespacedSecretName := types.NamespacedName{Namespace: namespace, Name: gitSource.Spec.SecretRef.Name}
	err := client.Get(context.TODO(), namespacedSecretName, coreSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch the secret object")
	}

	username := string(coreSecret.Data[corev1.BasicAuthUsernameKey])
	password := string(coreSecret.Data[corev1.BasicAuthPasswordKey])
	sshKey := string(coreSecret.Data[corev1.SSHAuthPrivateKey])
	var secret Secret
	if username != "" {
		secret = NewUsernamePassword(username, password)
	} else if password != "" {
		secret = NewOauthToken([]byte(password))
	} else if sshKey != "" {
		secret = NewSshKey([]byte(sshKey), coreSecret.Data["passphrase"])
	} else {
		return nil, fmt.Errorf("the provided secret does not contain any of the required parameters: [%s,%s,%s] or they are empty",
			corev1.BasicAuthUsernameKey, corev1.BasicAuthPasswordKey, corev1.SSHAuthPrivateKey)
	}
	return NewSecretProvider(secret), nil
}

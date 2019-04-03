package common_test

import (
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/git-service/pkg/controller/common"
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/redhat-developer/git-service/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

const pathToTestDir = "../../test"

func TestGetGitSecretWithNoCredentials(t *testing.T) {
	//given
	gs := test.NewGitSource()
	client, _ := test.PrepareClient(test.RegisterGvkObject(v1alpha1.SchemeGroupVersion, gs))

	//when
	secret, err := common.GetGitSecret(client, test.Namespace, gs)

	//then
	require.NoError(t, err)
	assert.Nil(t, secret)
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
		secret, err := common.GetGitSecret(client, test.Namespace, gs)

		//then
		require.NoError(t, err)
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
	secret, err := common.GetGitSecret(client, test.Namespace, gs)

	//then
	assert.Error(t, err)
	assert.Nil(t, secret)
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
		secret, err := common.GetGitSecret(client, test.Namespace, gs)

		//then
		require.NoError(t, err)
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
		secret, err := common.GetGitSecret(client, test.Namespace, gs)

		//then
		require.NoError(t, err)
		assert.Equal(t, git.SshKeyType, secret.SecretType())
		assert.NotEmpty(t, secret.SecretContent())
	}
}

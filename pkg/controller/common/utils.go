package common

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/git-service/pkg/git"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewNsdName(namespace, name string) types.NamespacedName {
	return types.NamespacedName{Namespace: namespace, Name: name}
}

func LogWithGSValues(log logr.Logger, gitSource *v1alpha1.GitSource, additional ...interface{}) logr.Logger {
	values := []interface{}{
		"name", gitSource.ObjectMeta.Name,
		"url", gitSource.Spec.URL,
		"ref", gitSource.Spec.Ref,
		"flavor", gitSource.Spec.Flavor,
	}
	values = append(values, additional...)

	return log.WithValues(values...)
}

func GetGitSecret(client client.Client, namespace string, gitSource *v1alpha1.GitSource) (git.Secret, error) {
	if gitSource.Spec.SecretRef == nil {
		return nil, nil
	}
	secret := &corev1.Secret{}
	namespacedSecretName := types.NamespacedName{Namespace: namespace, Name: gitSource.Spec.SecretRef.Name}
	err := client.Get(context.TODO(), namespacedSecretName, secret)
	if err != nil {
		return nil, err
	}

	username := string(secret.Data[corev1.BasicAuthUsernameKey])
	password := string(secret.Data[corev1.BasicAuthPasswordKey])
	sshKey := string(secret.Data[corev1.SSHAuthPrivateKey])
	if username != "" {
		return git.NewUsernamePassword(username, password), nil
	} else if password != "" {
		return git.NewOauthToken([]byte(password)), nil
	} else if sshKey != "" {
		return git.NewSshKey([]byte(sshKey), secret.Data["passphrase"]), nil
	}

	return nil, fmt.Errorf("the provided secret does not contain any of the required parameters: [%s,%s,%s] or they are empty",
		corev1.BasicAuthUsernameKey, corev1.BasicAuthPasswordKey, corev1.SSHAuthPrivateKey)
}

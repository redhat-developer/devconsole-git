package test

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func PrivateWithPassphrase(t *testing.T, pathToTestDir string) []byte {
	return readSecret(t, pathToTestDir, "ssh_with_passphrase/id_rsa")
}

func PublicWithPassphrase(t *testing.T, pathToTestDir string) []byte {
	return readSecret(t, pathToTestDir, "ssh_with_passphrase/id_rsa.pub")
}

func PrivateWithoutPassphrase(t *testing.T, pathToTestDir string) []byte {
	return readSecret(t, pathToTestDir, "ssh_without_passphrase/id_rsa")
}

func PublicWithoutPassphrase(t *testing.T, pathToTestDir string) []byte {
	return readSecret(t, pathToTestDir, "ssh_without_passphrase/id_rsa.pub")
}

func readSecret(t *testing.T, pathToTestDir string, secretPath string) []byte {
	content, err := ioutil.ReadFile(pathToTestDir + "/data/secret/" + secretPath)
	require.NoError(t, err)
	return content
}

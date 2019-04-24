package connection

import (
	"github.com/redhat-developer/devconsole-git/pkg/git"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository/bitbucket"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository/github"
	"github.com/redhat-developer/devconsole-git/pkg/git/repository/gitlab"
	"github.com/redhat-developer/devconsole-git/pkg/log"
	"github.com/redhat-developer/devconsole-git/pkg/test"
	"github.com/stretchr/testify/require"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

var logger = &log.GitSourceLogger{Logger: logf.Log}

func TestGenericGitUsingSshAccessingGitHub(t *testing.T) {
	// given
	ghSource := test.NewGitSource(test.WithURL("https://github.com/MatousJobanek/quarkus-knative.git"))
	allButGh := []repository.ServiceCreator{
		bitbucket.NewRepoServiceIfMatches(),
		gitlab.NewRepoServiceIfMatches(),
	}

	// when
	validationError := validateGitSourceWithSecret(logger, ghSource, git.NewSecretProvider(nil), allButGh)

	// then
	require.Nil(t, validationError)
}

func TestGenericGitUsingSshAccessingGitLab(t *testing.T) {
	// given
	ghSource := test.NewGitSource(test.WithURL("https://gitlab.com/matousjobanek/quarkus-knative.git"))
	allButGl := []repository.ServiceCreator{
		github.NewRepoServiceIfMatches(),
		bitbucket.NewRepoServiceIfMatches(),
	}

	// when
	validationError := validateGitSourceWithSecret(logger, ghSource, git.NewSecretProvider(nil), allButGl)

	// then
	require.Nil(t, validationError)
}

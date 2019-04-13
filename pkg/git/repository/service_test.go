package repository_test

import (
	"github.com/redhat-developer/devconsole-git/pkg/git/repository"
	"github.com/redhat-developer/devconsole-git/pkg/log"
	"github.com/redhat-developer/devconsole-git/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"testing"
)

var ghServiceCreator = test.NewDummyServiceCreator("github", false, test.S("github"), test.S())
var glServiceCreator = test.NewDummyServiceCreator("gitlab", false, test.S("gitlab"), test.S())
var bbServiceCreator = test.NewDummyServiceCreator("bitbucket", false, test.S("bitbucket"), test.S())
var failingServiceCreator = test.NewDummyServiceCreator("failing", true, test.S("failing"), test.S())
var logger = &log.GitSourceLogger{Logger: logf.Log}

func TestNewServiceReturnsCorrectService(t *testing.T) {
	// given
	creators := []repository.ServiceCreator{ghServiceCreator, glServiceCreator, bbServiceCreator}
	source := test.NewGitSource(test.WithFlavor("bitbucket"))

	// when
	service, err := repository.NewGitService(logger, source, nil, creators)

	// then
	require.NoError(t, err)
	require.NotNil(t, service)
	checker, _ := service.FileExistenceChecker()
	assert.Equal(t, "bitbucket", checker.GetListOfFoundFiles()[0])
}

func TestNewServiceReturnsNilService(t *testing.T) {
	// given
	creators := []repository.ServiceCreator{ghServiceCreator, glServiceCreator}
	source := test.NewGitSource(test.WithFlavor("bitbucket"))

	// when
	service, err := repository.NewGitService(logger, source, nil, creators)

	// then
	require.NoError(t, err)
	require.Nil(t, service)
}

func TestNewServiceReturnsError(t *testing.T) {
	// given
	creators := []repository.ServiceCreator{ghServiceCreator, glServiceCreator, failingServiceCreator, bbServiceCreator}
	source := test.NewGitSource(test.WithFlavor("bitbucket"))

	// when
	service, err := repository.NewGitService(logger, source, nil, creators)

	// then
	require.Error(t, err)
	require.Nil(t, service)
}

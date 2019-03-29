package repository_test

import (
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/redhat-developer/git-service/pkg/git/repository"
	"github.com/redhat-developer/git-service/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var ghServiceCreator = test.NewDummyServiceCreator("github", false, test.S("github"), test.S())
var glServiceCreator = test.NewDummyServiceCreator("gitlab", false, test.S("gitlab"), test.S())
var bbServiceCreator = test.NewDummyServiceCreator("bitbucket", false, test.S("bitbucket"), test.S())
var failingServiceCreator = test.NewDummyServiceCreator("failing", true, test.S("failing"), test.S())

func TestNewServiceReturnsCorrectService(t *testing.T) {
	// given
	creators := []repository.ServiceCreator{ghServiceCreator, glServiceCreator, bbServiceCreator}
	source := &git.Source{
		Flavor: "bitbucket",
	}

	// when
	service, err := repository.NewGitService(source, creators)

	// then
	require.NoError(t, err)
	require.NotNil(t, service)
	files, _ := service.GetListOfFilesInRootDir()
	assert.Equal(t, "bitbucket", files[0])
}

func TestNewServiceReturnsNilService(t *testing.T) {
	// given
	creators := []repository.ServiceCreator{ghServiceCreator, glServiceCreator}
	source := &git.Source{
		Flavor: "bitbucket",
	}

	// when
	service, err := repository.NewGitService(source, creators)

	// then
	require.NoError(t, err)
	require.Nil(t, service)
}

func TestNewServiceReturnsError(t *testing.T) {
	// given
	creators := []repository.ServiceCreator{ghServiceCreator, glServiceCreator, failingServiceCreator, bbServiceCreator}
	source := &git.Source{
		Flavor: "bitbucket",
	}

	// when
	service, err := repository.NewGitService(source, creators)

	// then
	require.Error(t, err)
	require.Nil(t, service)
}

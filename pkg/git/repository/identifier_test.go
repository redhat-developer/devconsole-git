package repository_test

import (
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/redhat-developer/git-service/pkg/git/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gittransport "gopkg.in/src-d/go-git.v4/plumbing/transport"
	"testing"
)

func TestNewStructuredIdentifierEmptyBranchAndStandardUrl(t *testing.T) {
	// given
	source := &git.Source{}
	endpoint, err := gittransport.NewEndpoint("https://github.com/fabric8-services/fabric8-tenant")
	require.NoError(t, err)

	// when
	identifier, err := repository.NewStructuredIdentifier(source, endpoint)

	// then
	require.NoError(t, err)
	assert.Equal(t, "fabric8-services", identifier.Owner)
	assert.Equal(t, "fabric8-tenant", identifier.Name)
	assert.Equal(t, "master", identifier.Branch)
}

func TestNewStructuredIdentifierDevBranchAndHttpsCloneUrl(t *testing.T) {
	// given
	source := &git.Source{
		Ref: "dev",
	}
	endpoint, err := gittransport.NewEndpoint("https://github.com/fabric8-services/fabric8-tenant.git")
	require.NoError(t, err)

	// when
	identifier, err := repository.NewStructuredIdentifier(source, endpoint)

	// then
	require.NoError(t, err)
	assert.Equal(t, "fabric8-services", identifier.Owner)
	assert.Equal(t, "fabric8-tenant", identifier.Name)
	assert.Equal(t, "dev", identifier.Branch)
}

func TestNewStructuredIdentifierEmptyBranchAndSshCloneUrl(t *testing.T) {
	// given
	source := &git.Source{}
	endpoint, err := gittransport.NewEndpoint("git@github.com:fabric8-services/fabric8-tenant.git")
	require.NoError(t, err)

	// when
	identifier, err := repository.NewStructuredIdentifier(source, endpoint)

	// then
	require.NoError(t, err)
	assert.Equal(t, "fabric8-services", identifier.Owner)
	assert.Equal(t, "fabric8-tenant", identifier.Name)
	assert.Equal(t, "master", identifier.Branch)
}

func TestNewStructuredIdentifierEmptyBranchAndSshClonePath(t *testing.T) {
	// given
	source := &git.Source{}
	endpoint, err := gittransport.NewEndpoint("git@github.com:/fabric8-services/fabric8-tenant.git")
	require.NoError(t, err)

	// when
	identifier, err := repository.NewStructuredIdentifier(source, endpoint)

	// then
	require.NoError(t, err)
	assert.Equal(t, "fabric8-services", identifier.Owner)
	assert.Equal(t, "fabric8-tenant", identifier.Name)
	assert.Equal(t, "master", identifier.Branch)
}

func TestNewStructuredIdentifierEmptyBranchAndWrongPathInHttpsUrl(t *testing.T) {
	// given
	source := &git.Source{}
	endpoint, err := gittransport.NewEndpoint("https://github.com/fabric8-services-fabric8-tenant")
	require.NoError(t, err)

	// when
	_, err = repository.NewStructuredIdentifier(source, endpoint)

	// then
	require.Error(t, err)
}

func TestNewStructuredIdentifierEmptyBranchAndWrongPathInSshUrl(t *testing.T) {
	// given
	source := &git.Source{}
	endpoint, err := gittransport.NewEndpoint("git@github.com:/fabric8-services-fabric8-tenant")
	require.NoError(t, err)

	// when
	_, err = repository.NewStructuredIdentifier(source, endpoint)

	// then
	require.Error(t, err)
}

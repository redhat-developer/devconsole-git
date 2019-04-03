package bitbucket

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"

	gittransport "gopkg.in/src-d/go-git.v4/plumbing/transport"
)

func TestGetBaseURLReturnsDefault(t *testing.T) {
	// given
	endpoint, err := gittransport.NewEndpoint("https://bitbucket.org/some-org/some-repo/")
	require.NoError(t, err)

	// when
	url := getBaseURL(endpoint)

	// then
	assert.Equal(t, "https://api.bitbucket.org/", url)
}

func TestGetBaseURLReturnsCustom(t *testing.T) {
	// given
	endpoint, err := gittransport.NewEndpoint("https://bitbucket.redhat.com/some-org/some-repo/")
	require.NoError(t, err)

	// when
	url := getBaseURL(endpoint)

	// then
	assert.Equal(t, "https://api.bitbucket.redhat.com/", url)
}

func TestGetBaseURLFromDefaultSsh(t *testing.T) {
	// given
	endpoint, err := gittransport.NewEndpoint("git@bitbucket.org:some-org/some-repo.git")
	require.NoError(t, err)

	// when
	url := getBaseURL(endpoint)

	// then
	assert.Equal(t, "https://api.bitbucket.org/", url)
}

func TestGetBaseURLFromCustomSsh(t *testing.T) {
	// given
	endpoint, err := gittransport.NewEndpoint("git@bitbucket.redhat.com:some-org/some-repo.git")
	require.NoError(t, err)

	// when
	url := getBaseURL(endpoint)

	// then
	assert.Equal(t, "https://api.bitbucket.redhat.com/", url)
}

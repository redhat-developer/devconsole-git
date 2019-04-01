package detector

import (
	"fmt"
	"github.com/redhat-developer/git-service/pkg/test"
	"io/ioutil"
	"os"
	"testing"

	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/stretchr/testify/require"
)

var (
	homeDir  = os.Getenv("HOME")
	nilSlice = func() []string {
		return nil
	}
	failingService          = test.NewDummyService("failing", true, test.S("failing"), test.S())
	failingFilesService     = test.NewDummyService("failing-files", false, nilSlice, test.S("Java"))
	failingLanguagesService = test.NewDummyService("failing-languages", false, test.S("any"), nilSlice)
)

func TestFailingCreator(t *testing.T) {
	// given
	source := test.NewGitSource(test.WithFlavor(failingService.Flavor))

	// when
	buildEnvStats, err := detectBuildEnvs(source, nil, append(allCreators, failingService.Creator()))

	// then
	require.Error(t, err)
	require.Contains(t, err.Error(), "creation failed")
	require.Nil(t, buildEnvStats)
}

func TestFailingGetFileList(t *testing.T) {
	// when
	buildEnvStats, err := detectBuildEnvsUsingService(failingFilesService)

	// then
	require.Error(t, err)
	require.Equal(t, err.Error(), "failing files")
	require.Nil(t, buildEnvStats)
}

func TestFailingGetLanguagesList(t *testing.T) {
	// when
	buildEnvStats, err := detectBuildEnvsUsingService(failingLanguagesService)

	// then
	require.Error(t, err)
	require.Equal(t, err.Error(), "failing languages")
	require.Nil(t, buildEnvStats)
}

// ignored tests as they reach the real services

func XTestGitHubDetectorWithToken(t *testing.T) {
	token, err := ioutil.ReadFile(homeDir + "/.github-auth")
	require.NoError(t, err)

	ghSource := test.NewGitSource(test.WithURL("https://github.com/wildfly/wildfly"))

	buildEnvStats, err := DetectBuildEnvironments(ghSource, git.NewOauthToken(token))
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func XTestGitHubDetectorWithUsernameAndPassword(t *testing.T) {
	ghSource := test.NewGitSource(test.WithURL("https://github.com/wildfly/wildfly"))

	buildEnvStats, err := DetectBuildEnvironments(ghSource, git.NewUsernamePassword("anonymous", ""))
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func XTestGitLabDetectorWithToken(t *testing.T) {

	glSource := test.NewGitSource(test.WithURL("https://gitlab.com/gitlab-org/gitlab-qa"))

	buildEnvStats, err := DetectBuildEnvironments(glSource, git.NewOauthToken([]byte("")))
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func XTestGitLabDetectorWithUsernamePassword(t *testing.T) {

	glSource := test.NewGitSource(test.WithURL("https://gitlab.com/gitlab-org/gitlab-qa"))

	buildEnvStats, err := DetectBuildEnvironments(glSource, git.NewUsernamePassword("", ""))
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func printBuildEnvStats(buildEnvStats *BuildEnvStats) {
	fmt.Println(buildEnvStats.SortedLanguages)
	for _, build := range buildEnvStats.DetectedBuildTools {
		fmt.Println(*build)
	}
}

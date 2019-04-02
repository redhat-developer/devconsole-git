package detector

import (
	"fmt"
	"github.com/redhat-developer/git-service/pkg/test"
	"github.com/stretchr/testify/assert"
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

func TestBitbucketDetectorWithToken(t *testing.T) {
	// given
	glSource := test.NewGitSource(test.WithURL("https://bitbucket.org/mjobanek-rh/quarkus-knative"))

	// when
	buildEnvStats, err := DetectBuildEnvironments(glSource, git.NewOauthToken([]byte("")))

	// then
	require.NoError(t, err)
	require.Len(t, buildEnvStats.DetectedBuildTools, 1)
	assertContainsBuildTool(t, buildEnvStats.DetectedBuildTools, Maven, "pom.xml")
	require.Len(t, buildEnvStats.SortedLanguages, 1)
	assert.Equal(t, "java", buildEnvStats.SortedLanguages[0])
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

func XTestBitbucketDetectorWithUsernameAndPassword(t *testing.T) {
	ghSource := test.NewGitSource(test.WithURL("https://bitbucket.org/atlassian/asap-java"))

	buildEnvStats, err := DetectBuildEnvironments(ghSource, git.NewUsernamePassword("", ""))
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func XTestBitbucketDetectorWithToken(t *testing.T) {
	ghSource := test.NewGitSource(test.WithURL("https://bitbucket.org/atlassian/asap-java"))

	buildEnvStats, err := DetectBuildEnvironments(ghSource, git.NewOauthToken([]byte("")))
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func printBuildEnvStats(buildEnvStats *BuildEnvStats) {
	fmt.Println(buildEnvStats.SortedLanguages)
	for _, build := range buildEnvStats.DetectedBuildTools {
		fmt.Println(*build)
	}
}

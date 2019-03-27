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
	source := &git.Source{Flavor: failingService.Flavor}

	// when
	buildEnvStats, err := detectBuildEnvs(source, append(allCreators, failingService.Creator()))

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

	ghSource := &git.Source{
		URL:    "https://github.com/wildfly/wildfly",
		Secret: git.NewOauthToken(token),
	}

	buildEnvStats, err := DetectBuildEnvironments(ghSource)
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func XTestGitHubDetectorWithUsernameAndPassword(t *testing.T) {
	ghSource := &git.Source{
		URL:    "https://github.com/wildfly/wildfly",
		Secret: git.NewUsernamePassword("anonymous", ""),
	}

	buildEnvStats, err := DetectBuildEnvironments(ghSource)
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func printBuildEnvStats(buildEnvStats *BuildEnvStats) {
	fmt.Println(buildEnvStats.SortedLanguages)
	for _, build := range buildEnvStats.DetectedBuildTools {
		fmt.Println(*build)
	}
}

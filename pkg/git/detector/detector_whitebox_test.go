package detector

import (
	"fmt"
	"github.com/redhat-developer/git-service/pkg/git/repository"
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

const pathToTestDir = "../../test"

func TestDetectBuildEnvsShouldUseGenericGitIfNotOtherMatches(t *testing.T) {
	// given
	dummyRepo := test.NewDummyGitRepo(t, repository.Master)
	dummyRepo.Commit(
		"pom.xml", "package.json", "other.json", "src/main/java/Any.java", "src/main/java/Another.java",
		"src/main/java/Third.java", "pkg/main.go", "pkg/cool.go", "pkg/cool_test.go", "pkg/another.go")

	sshKey := git.NewSshKey(test.PrivateWithoutPassphrase(t, pathToTestDir), []byte(""))
	source := test.NewGitSource(test.WithURL(dummyRepo.Path), test.WithFlavor("not-existing"))

	// when
	buildEnvStats, err := detectBuildEnvs(source, sshKey, allCreators)

	// then
	require.NoError(t, err)
	require.NotNil(t, buildEnvStats)

	buildTools := buildEnvStats.DetectedBuildTools
	require.Len(t, buildTools, 2)

	assertContainsBuildTool(t, buildTools, Maven, "pom.xml")
	assertContainsBuildTool(t, buildTools, NodeJS, "package.json")

	langs := buildEnvStats.SortedLanguages
	assert.Len(t, langs, 4)
	assert.Equal(t, "Go", langs[0])
	assert.Equal(t, "Java", langs[1])
	assert.Equal(t, "JSON", langs[2])
	assert.Equal(t, "XML", langs[3])
}

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

func TestFailingGenericGitCreation(t *testing.T) {
	// given
	dummyRepo := test.NewDummyGitRepo(t, repository.Master)
	sshKey := git.NewSshKey(test.PrivateWithPassphrase(t, pathToTestDir), []byte(""))
	source := test.NewGitSource(test.WithURL(dummyRepo.Path), test.WithFlavor("not-existing"))

	// when
	buildEnvStats, err := detectBuildEnvs(source, sshKey, allCreators)

	// then
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot decode encrypted private keys")
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

func XTestGenericGitUsingSshAccessingGitHub(t *testing.T) {

	buffer, err := ioutil.ReadFile(homeDir + "/.ssh/id_rsa")
	require.NoError(t, err)

	ghSource := test.NewGitSource(test.WithURL("git@github.com:wildfly/wildfly.git"))

	buildEnvStats, err := DetectBuildEnvironments(ghSource, git.NewSshKey(buffer, []byte("passphrase")))
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func XTestGitLabDetectorWithToken(t *testing.T) {

	glSource := test.NewGitSource(test.WithURL("https://gitlab.com/gitlab-org/gitlab-qa"))

	buildEnvStats, err := DetectBuildEnvironments(glSource, git.NewOauthToken([]byte("")))
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func XTestGenericGitUsingSshAccessingGitLab(t *testing.T) {

	buffer, err := ioutil.ReadFile(homeDir + "/.ssh/id_rsa")
	require.NoError(t, err)

	ghSource := test.NewGitSource(test.WithURL("git@gitlab.cee.redhat.com:mjobanek/housekeeping.git"))

	buildEnvStats, err := DetectBuildEnvironments(ghSource, git.NewSshKey(buffer, []byte("passphrase")))
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func printBuildEnvStats(buildEnvStats *BuildEnvStats) {
	fmt.Println(buildEnvStats.SortedLanguages)
	for _, build := range buildEnvStats.DetectedBuildTools {
		fmt.Println(*build)
	}
}

package detector

import (
	"fmt"
	"github.com/redhat-developer/devconsole-api/pkg/apis/devconsole/v1alpha1"
	"github.com/redhat-developer/git-service/pkg/git/detector/build"
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
	failingService          = test.NewDummyService("failing", true, test.S("failing"), test.S(), true)
	failingFilesService     = test.NewDummyService("failing-files", false, nilSlice, test.S("Java"), true)
	failingLanguagesService = test.NewDummyService("failing-languages", false, test.S("any"), nilSlice, true)
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
	buildEnvStats, err := detectBuildEnvs(source, git.NewSecretProvider(sshKey), allEnvServiceCreators(true))

	// then
	require.NoError(t, err)
	require.NotNil(t, buildEnvStats)

	buildTools := buildEnvStats.DetectedBuildTypes
	require.Len(t, buildTools, 2)

	assertContainsBuildTool(t, buildTools, build.Maven, "pom.xml")
	assertContainsBuildTool(t, buildTools, build.NodeJS, "package.json")

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
	buildEnvStats, err := detectBuildEnvs(source, nil, append(allEnvServiceCreators(true), failingService.Creator()))

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
	buildEnvStats, err := detectBuildEnvs(source, git.NewSecretProvider(sshKey), allEnvServiceCreators(true))

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

func TestBitbucketDetectorWithDefault(t *testing.T) {
	// given
	glSource := test.NewGitSource(test.WithURL("https://bitbucket.org/mjobanek-rh/quarkus-knative"))

	// when
	buildEnvStats, err := DetectBuildEnvironments(glSource)

	// then
	require.NoError(t, err)
	require.Len(t, buildEnvStats.DetectedBuildTypes, 1)
	assertContainsBuildTool(t, buildEnvStats.DetectedBuildTypes, build.Maven, "pom.xml")
	require.Len(t, buildEnvStats.SortedLanguages, 1)
	assert.Equal(t, "java", buildEnvStats.SortedLanguages[0])
}

func TestGitLabDetectorWithDefault(t *testing.T) {
	// given
	glSource := test.NewGitSource(test.WithURL("https://gitlab.com/matousjobanek/quarkus-knative"))

	// when
	buildEnvStats, err := DetectBuildEnvironments(glSource)

	// then
	require.NoError(t, err)
	require.Len(t, buildEnvStats.DetectedBuildTypes, 1)
	assertContainsBuildTool(t, buildEnvStats.DetectedBuildTypes, build.Maven, "pom.xml")
	require.Len(t, buildEnvStats.SortedLanguages, 1)
	assert.Equal(t, "Java", buildEnvStats.SortedLanguages[0])
}

func TestGitHubDetectorWithDefault(t *testing.T) {
	// given
	glSource := test.NewGitSource(test.WithURL("https://github.com/MatousJobanek/quarkus-knative"))

	// when
	buildEnvStats, err := DetectBuildEnvironments(glSource)

	// then
	require.NoError(t, err)
	require.Len(t, buildEnvStats.DetectedBuildTypes, 1)
	assertContainsBuildTool(t, buildEnvStats.DetectedBuildTypes, build.Maven, "pom.xml")
}

func TestGenericGitUsingSshAccessingGitLabWithDefaultCredentials(t *testing.T) {
	// given
	glSource := test.NewGitSource(test.WithURL("https://gitlab.com/matousjobanek/quarkus-knative"))

	// when
	buildEnvStats, err := detectBuildEnvs(glSource, git.NewSecretProvider(nil), []repository.ServiceCreator{})

	// then
	require.NoError(t, err)
	require.Len(t, buildEnvStats.DetectedBuildTypes, 1)
	assertContainsBuildTool(t, buildEnvStats.DetectedBuildTypes, build.Maven, "pom.xml")
	require.Len(t, buildEnvStats.SortedLanguages, 6)
	assert.Contains(t, buildEnvStats.SortedLanguages, "Java")
}

func TestGenericGitUsingSshAccessingBitbucketWithDefaultCredentials(t *testing.T) {
	// given
	glSource := test.NewGitSource(test.WithURL("https://bitbucket.org/mjobanek-rh/quarkus-knative"))

	// when
	buildEnvStats, err := detectBuildEnvs(glSource, git.NewSecretProvider(nil), []repository.ServiceCreator{})

	// then
	require.NoError(t, err)
	require.Len(t, buildEnvStats.DetectedBuildTypes, 1)
	assertContainsBuildTool(t, buildEnvStats.DetectedBuildTypes, build.Maven, "pom.xml")
	require.Len(t, buildEnvStats.SortedLanguages, 6)
	assert.Contains(t, buildEnvStats.SortedLanguages, "Java")
}

func TestGenericGitUsingSshAccessingGitHubWithDefaultCredentials(t *testing.T) {
	// given
	ghSource := test.NewGitSource(test.WithURL("https://github.com/MatousJobanek/quarkus-knative"))

	// when
	buildEnvStats, err := detectBuildEnvs(ghSource, git.NewSecretProvider(nil), []repository.ServiceCreator{})

	// then
	require.NoError(t, err)
	require.Len(t, buildEnvStats.DetectedBuildTypes, 1)
	assertContainsBuildTool(t, buildEnvStats.DetectedBuildTypes, build.Maven, "pom.xml")
	require.Len(t, buildEnvStats.SortedLanguages, 6)
	assert.Contains(t, buildEnvStats.SortedLanguages, "Java")
}

// ignored tests as they reach the real services or needs specific credentials

func TestGitHubDetectorWithToken(t *testing.T) {
	t.Skip("skip to avoid API rate limits")

	token, err := ioutil.ReadFile(homeDir + "/.github-auth")
	require.NoError(t, err)

	ghSource := test.NewGitSource(test.WithURL("https://github.com/wildfly/wildfly"))

	buildEnvStats, err := DetectBuildEnvironmentsWithSecret(ghSource, git.NewOauthToken(token))
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func TestGitHubDetectorWithUsernameAndPassword(t *testing.T) {
	t.Skip("skip to avoid API rate limits")

	ghSource := test.NewGitSource(test.WithURL("https://github.com/wildfly/wildfly"))

	buildEnvStats, err := DetectBuildEnvironmentsWithSecret(ghSource, git.NewUsernamePassword("anonymous", ""))
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func TestGenericGitUsingSshAccessingGitHub(t *testing.T) {
	t.Skip("skip as it depends on ssh key registered in GH")

	buffer, err := ioutil.ReadFile(homeDir + "/.ssh/id_rsa")
	require.NoError(t, err)

	ghSource := test.NewGitSource(test.WithURL("git@github.com:wildfly/wildfly.git"))

	buildEnvStats, err := DetectBuildEnvironmentsWithSecret(ghSource, git.NewSshKey(buffer, []byte("passphrase")))
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func TestGenericGitUsingSshAccessingGitLab(t *testing.T) {
	t.Skip("skip as it depends on ssh key registered in GH")

	buffer, err := ioutil.ReadFile(homeDir + "/.ssh/id_rsa")
	require.NoError(t, err)

	ghSource := test.NewGitSource(test.WithURL("git@gitlab.cee.redhat.com:mjobanek/housekeeping.git"))

	buildEnvStats, err := DetectBuildEnvironmentsWithSecret(ghSource, git.NewSshKey(buffer, []byte("passphrase")))
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func TestBitbucketDetectorWithUsernameAndPassword(t *testing.T) {
	t.Skip("skip as Bitbucket doesn't support authN with empty credential nor with anonymous user")

	ghSource := test.NewGitSource(test.WithURL("https://bitbucket.org/atlassian/asap-java"))

	buildEnvStats, err := DetectBuildEnvironmentsWithSecret(ghSource, git.NewUsernamePassword("", ""))

	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func TestGitLabDetectorWithUsernamePassword(t *testing.T) {
	t.Skip("skip as GitLab doesn't support authN with empty credential nor with anonymous user")

	glSource := test.NewGitSource(test.WithURL("https://gitlab.com/gitlab-org/gitlab-qa"))

	buildEnvStats, err := DetectBuildEnvironmentsWithSecret(glSource, git.NewUsernamePassword("", ""))
	require.NoError(t, err)
	printBuildEnvStats(buildEnvStats)
}

func printBuildEnvStats(buildEnvStats *v1alpha1.BuildEnvStats) {
	fmt.Println(buildEnvStats.SortedLanguages)
	for _, build := range buildEnvStats.DetectedBuildTypes {
		fmt.Println(build)
	}
}

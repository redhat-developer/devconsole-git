package detector

import (
	"fmt"
	"github.com/redhat-developer/git-service/pkg/git"
	"github.com/redhat-developer/git-service/pkg/git/repository"
	"github.com/redhat-developer/git-service/pkg/test"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	javaService = test.NewDummyService("java", false,
		test.S("somefile", "gradlew", "pom.xml"), test.S("Java"))

	javaAndGoService = test.NewDummyService("javaAndGo", false,
		test.S("somefile", "main.go", "pom.xml"), test.S("Java", "Go"))

	rubyService = test.NewDummyService("ruby", false,
		test.S("somefile", "Gemfile"), test.S("Ruby"))

	nodeJsService = test.NewDummyService("nodejs", false,
		test.S("somefile", "package.json", "gulpfile.js"), test.S("NodeJS"))

	phpService = test.NewDummyService("php", false,
		test.S("somefile", "index.php"), test.S("PHP"))

	pythonService = test.NewDummyService("python", false,
		test.S("somefile", "requirements.txt"), test.S("Python"))

	perlService = test.NewDummyService("perl", false,
		test.S("somefile", "index.pl"), test.S("Perl"))

	dotnetService = test.NewDummyService("dotnet", false,
		test.S("somefile", "project.json", "some.csproj"), test.S("C#"))

	unknownEnvService = test.NewDummyService("unknown", false,
		test.S("somefile", "another.file", "cool.file"), test.S("CoolLang"))

	allEnvServices = []*test.DummyService{javaService, javaAndGoService, rubyService,
		nodeJsService, phpService, pythonService, perlService, dotnetService}
	allCreators = allEnvServiceCreators()
)

func TestDetectJavaMavenAndGradle(t *testing.T) {
	// given
	source := &git.Source{Flavor: javaService.Flavor}

	// when
	buildEnvStats, err := detectBuildEnvs(source, allCreators)

	// then
	require.NoError(t, err)
	require.NotNil(t, buildEnvStats)

	buildTools := buildEnvStats.DetectedBuildTools
	require.Len(t, buildTools, 2)

	assertContainsBuildTool(t, buildTools, Gradle, "gradlew")
	assertContainsBuildTool(t, buildTools, Maven, "pom.xml")

	langs := buildEnvStats.SortedLanguages
	assert.Len(t, langs, 1)
	assert.Equal(t, "Java", langs[0])
}

func TestDetectJavaMavenAndGo(t *testing.T) {
	// given
	source := &git.Source{Flavor: javaAndGoService.Flavor}

	// when
	buildEnvStats, err := detectBuildEnvs(source, allCreators)

	// then
	require.NoError(t, err)
	require.NotNil(t, buildEnvStats)

	buildTools := buildEnvStats.DetectedBuildTools
	require.Len(t, buildTools, 2)

	assertContainsBuildTool(t, buildTools, Golang, "main.go")
	assertContainsBuildTool(t, buildTools, Maven, "pom.xml")

	langs := buildEnvStats.SortedLanguages
	assert.Len(t, langs, 2)
	assert.Contains(t, langs, "Java")
	assert.Contains(t, langs, "Go")
}

func TestDetectRuby(t *testing.T) {
	// given
	source := &git.Source{Flavor: rubyService.Flavor}

	// when
	buildEnvStats, err := detectBuildEnvs(source, allCreators)

	// then
	require.NoError(t, err)
	require.NotNil(t, buildEnvStats)

	buildTools := buildEnvStats.DetectedBuildTools
	require.Len(t, buildTools, 1)

	assertContainsBuildTool(t, buildTools, Ruby, "Gemfile")

	langs := buildEnvStats.SortedLanguages
	assert.Len(t, langs, 1)
	assert.Equal(t, "Ruby", langs[0])
}

func TestDetectPHP(t *testing.T) {
	// given
	source := &git.Source{Flavor: phpService.Flavor}

	// when
	buildEnvStats, err := detectBuildEnvs(source, allCreators)

	// then
	require.NoError(t, err)
	require.NotNil(t, buildEnvStats)

	buildTools := buildEnvStats.DetectedBuildTools
	require.Len(t, buildTools, 1)

	assertContainsBuildTool(t, buildTools, PHP, "index.php")

	langs := buildEnvStats.SortedLanguages
	assert.Len(t, langs, 1)
	assert.Equal(t, "PHP", langs[0])
}

func TestDetectNodeJS(t *testing.T) {
	// given
	source := &git.Source{Flavor: nodeJsService.Flavor}

	// when
	buildEnvStats, err := detectBuildEnvs(source, allCreators)

	// then
	require.NoError(t, err)
	require.NotNil(t, buildEnvStats)

	buildTools := buildEnvStats.DetectedBuildTools
	require.Len(t, buildTools, 1)

	assertContainsBuildTool(t, buildTools, NodeJS, "package.json", "gulpfile.js")

	langs := buildEnvStats.SortedLanguages
	assert.Len(t, langs, 1)
	assert.Equal(t, "NodeJS", langs[0])
}

func TestDetectPython(t *testing.T) {
	// given
	source := &git.Source{Flavor: pythonService.Flavor}

	// when
	buildEnvStats, err := detectBuildEnvs(source, allCreators)

	// then
	require.NoError(t, err)
	require.NotNil(t, buildEnvStats)

	buildTools := buildEnvStats.DetectedBuildTools
	require.Len(t, buildTools, 1)

	assertContainsBuildTool(t, buildTools, Python, "requirements.txt")

	langs := buildEnvStats.SortedLanguages
	assert.Len(t, langs, 1)
	assert.Equal(t, "Python", langs[0])
}

func TestDetectPerl(t *testing.T) {
	// given
	source := &git.Source{Flavor: perlService.Flavor}

	// when
	buildEnvStats, err := detectBuildEnvs(source, allCreators)

	// then
	require.NoError(t, err)
	require.NotNil(t, buildEnvStats)

	buildTools := buildEnvStats.DetectedBuildTools
	require.Len(t, buildTools, 1)

	assertContainsBuildTool(t, buildTools, Perl, "index.pl")

	langs := buildEnvStats.SortedLanguages
	assert.Len(t, langs, 1)
	assert.Equal(t, "Perl", langs[0])
}

func TestDetectDotnet(t *testing.T) {
	// given
	source := &git.Source{Flavor: dotnetService.Flavor}

	// when
	buildEnvStats, err := detectBuildEnvs(source, allCreators)

	// then
	require.NoError(t, err)
	require.NotNil(t, buildEnvStats)

	buildTools := buildEnvStats.DetectedBuildTools
	require.Len(t, buildTools, 1)

	assertContainsBuildTool(t, buildTools, Dotnet, "project.json", "some.csproj")

	langs := buildEnvStats.SortedLanguages
	assert.Len(t, langs, 1)
	assert.Equal(t, "C#", langs[0])
}

func assertContainsBuildTool(t *testing.T, detected []*DetectedBuildTool, expectedBuildTool BuildTool, files ...string) {
	var found bool
	for _, tool := range detected {
		if tool.name == expectedBuildTool.Name {
			assert.Equal(t, expectedBuildTool.Language, tool.language)
			assert.Len(t, tool.detectedFiles, len(files))
			for _, file := range files {
				assert.Contains(t, tool.detectedFiles, file)
			}
			found = true
			break
		}
	}
	if !found {
		assert.Fail(t, fmt.Sprintf("the list %v does not contain tool with name %s, language %s and files %v",
			detected, expectedBuildTool.Name, expectedBuildTool.Language, files))
	}
}

func allEnvServiceCreators() []repository.ServiceCreator {
	var creators []repository.ServiceCreator
	for _, service := range allEnvServices {
		creators = append(creators, service.Creator())
	}
	return creators
}

package detector

import (
	"github.com/redhat-developer/git-service/pkg/git/detector/build"
	"github.com/redhat-developer/git-service/pkg/git/repository"
	"github.com/redhat-developer/git-service/pkg/log"
	"github.com/redhat-developer/git-service/pkg/test"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
	"testing"

	"github.com/stretchr/testify/require"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	javaServices = test.NewDummyServices("java", false,
		test.S("somefile", "gradlew", "pom.xml"), test.S("Java"))

	javaAndGoServices = test.NewDummyServices("javaAndGo", false,
		test.S("somefile", "main.go", "pom.xml"), test.S("Java", "Go"))

	rubyServices = test.NewDummyServices("ruby", false,
		test.S("somefile", "Gemfile"), test.S("Ruby"))

	nodeJsServices = test.NewDummyServices("nodejs", false,
		test.S("somefile", "package.json", "gulpfile.js"), test.S("NodeJS"))

	phpServices = test.NewDummyServices("php", false,
		test.S("somefile", "index.php"), test.S("PHP"))

	pythonServices = test.NewDummyServices("python", false,
		test.S("somefile", "requirements.txt"), test.S("Python"))

	perlServices = test.NewDummyServices("perl", false,
		test.S("somefile", "index.pl"), test.S("Perl"))

	dotnetServices = test.NewDummyServices("dotnet", false,
		test.S("somefile", "project.json", "app.csproj"), test.S("C#"))

	unknownEnvServices = test.NewDummyServices("unknown", false,
		test.S("somefile", "another.file", "cool.file"), test.S("CoolLang"))

	allEnvServices = [][]*test.DummyService{javaServices, javaAndGoServices, rubyServices,
		nodeJsServices, phpServices, pythonServices, perlServices, dotnetServices}
)
var logger = &log.GitSourceLogger{Logger: logf.Log}

func TestDetectJavaMavenAndGradle(t *testing.T) {
	defer gock.OffAll()
	for _, service := range javaServices {
		// given
		source := test.NewGitSource(test.WithFlavor(service.Flavor))

		// when
		buildEnvStats, err := detectBuildEnvs(logger, source, nil, allEnvServiceCreators(service.UseFilesChecker))

		// then
		require.NoError(t, err)
		require.NotNil(t, buildEnvStats)

		buildTools := buildEnvStats.DetectedBuildTypes
		require.Len(t, buildTools, 2)

		assertContainsBuildTool(t, buildTools, build.Gradle, "gradlew")
		assertContainsBuildTool(t, buildTools, build.Maven, "pom.xml")

		langs := buildEnvStats.SortedLanguages
		assert.Len(t, langs, 1)
		assert.Equal(t, "Java", langs[0])
	}
}

func TestDetectJavaMavenAndGo(t *testing.T) {
	defer gock.OffAll()
	for _, service := range javaAndGoServices {
		// given
		source := test.NewGitSource(test.WithFlavor(service.Flavor))

		// when
		buildEnvStats, err := detectBuildEnvs(logger, source, nil, allEnvServiceCreators(service.UseFilesChecker))

		// then
		require.NoError(t, err)
		require.NotNil(t, buildEnvStats)

		buildTools := buildEnvStats.DetectedBuildTypes
		require.Len(t, buildTools, 2)

		assertContainsBuildTool(t, buildTools, build.Golang, "main.go")
		assertContainsBuildTool(t, buildTools, build.Maven, "pom.xml")

		langs := buildEnvStats.SortedLanguages
		assert.Len(t, langs, 2)
		assert.Contains(t, langs, "Java")
		assert.Contains(t, langs, "Go")
	}
}

func TestDetectRuby(t *testing.T) {
	defer gock.OffAll()
	for _, service := range rubyServices {
		// given
		source := test.NewGitSource(test.WithFlavor(service.Flavor))

		// when
		buildEnvStats, err := detectBuildEnvs(logger, source, nil, allEnvServiceCreators(service.UseFilesChecker))

		// then
		require.NoError(t, err)
		require.NotNil(t, buildEnvStats)

		buildTools := buildEnvStats.DetectedBuildTypes
		require.Len(t, buildTools, 1)

		assertContainsBuildTool(t, buildTools, build.Ruby, "Gemfile")

		langs := buildEnvStats.SortedLanguages
		assert.Len(t, langs, 1)
		assert.Equal(t, "Ruby", langs[0])
	}
}

func TestDetectPHP(t *testing.T) {
	defer gock.OffAll()
	for _, service := range phpServices {
		// given
		source := test.NewGitSource(test.WithFlavor(service.Flavor))

		// when
		buildEnvStats, err := detectBuildEnvs(logger, source, nil, allEnvServiceCreators(service.UseFilesChecker))

		// then
		require.NoError(t, err)
		require.NotNil(t, buildEnvStats)

		buildTools := buildEnvStats.DetectedBuildTypes
		require.Len(t, buildTools, 1)

		assertContainsBuildTool(t, buildTools, build.PHP, "index.php")

		langs := buildEnvStats.SortedLanguages
		assert.Len(t, langs, 1)
		assert.Equal(t, "PHP", langs[0])
	}
}

func TestDetectNodeJS(t *testing.T) {
	defer gock.OffAll()
	for _, service := range nodeJsServices {
		// given
		source := test.NewGitSource(test.WithFlavor(service.Flavor))

		// when
		buildEnvStats, err := detectBuildEnvs(logger, source, nil, allEnvServiceCreators(service.UseFilesChecker))

		// then
		require.NoError(t, err)
		require.NotNil(t, buildEnvStats)

		buildTools := buildEnvStats.DetectedBuildTypes
		require.Len(t, buildTools, 1)

		assertContainsBuildTool(t, buildTools, build.NodeJS, "package.json", "gulpfile.js")

		langs := buildEnvStats.SortedLanguages
		assert.Len(t, langs, 1)
		assert.Equal(t, "NodeJS", langs[0])
	}
}

func TestDetectPython(t *testing.T) {
	defer gock.OffAll()
	for _, service := range pythonServices {
		// given
		source := test.NewGitSource(test.WithFlavor(service.Flavor))

		// when
		buildEnvStats, err := detectBuildEnvs(logger, source, nil, allEnvServiceCreators(service.UseFilesChecker))

		// then
		require.NoError(t, err)
		require.NotNil(t, buildEnvStats)

		buildTools := buildEnvStats.DetectedBuildTypes
		require.Len(t, buildTools, 1)

		assertContainsBuildTool(t, buildTools, build.Python, "requirements.txt")

		langs := buildEnvStats.SortedLanguages
		assert.Len(t, langs, 1)
		assert.Equal(t, "Python", langs[0])
	}
}

func TestDetectPerl(t *testing.T) {
	defer gock.OffAll()
	for _, service := range perlServices {
		// given
		source := test.NewGitSource(test.WithFlavor(service.Flavor))

		// when
		buildEnvStats, err := detectBuildEnvs(logger, source, nil, allEnvServiceCreators(service.UseFilesChecker))

		// then
		require.NoError(t, err)
		require.NotNil(t, buildEnvStats)

		buildTools := buildEnvStats.DetectedBuildTypes
		require.Len(t, buildTools, 1)

		assertContainsBuildTool(t, buildTools, build.Perl, "index.pl")

		langs := buildEnvStats.SortedLanguages
		assert.Len(t, langs, 1)
		assert.Equal(t, "Perl", langs[0])
	}
}

func TestDetectDotnet(t *testing.T) {
	defer gock.OffAll()
	for _, service := range dotnetServices {
		// given
		source := test.NewGitSource(test.WithFlavor(service.Flavor))

		// when
		buildEnvStats, err := detectBuildEnvs(logger, source, nil, allEnvServiceCreators(service.UseFilesChecker))

		// then
		require.NoError(t, err)
		require.NotNil(t, buildEnvStats)

		buildTools := buildEnvStats.DetectedBuildTypes
		require.Len(t, buildTools, 1)

		assertContainsBuildTool(t, buildTools, build.Dotnet, "project.json", "app.csproj")

		langs := buildEnvStats.SortedLanguages
		assert.Len(t, langs, 1)
		assert.Equal(t, "C#", langs[0])
	}
}

func allEnvServiceCreators(useFilesChecker bool) []repository.ServiceCreator {
	var creators []repository.ServiceCreator
	for _, services := range allEnvServices {
		for _, service := range services {
			if (useFilesChecker && service.UseFilesChecker) || (!useFilesChecker && !service.UseFilesChecker) {
				creators = append(creators, service.Creator())
			}
		}
	}
	return creators
}
